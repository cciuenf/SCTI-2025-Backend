package services

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"scti/config"
	"scti/internal/models"
	repos "scti/internal/repositories"
	"strings"
	"text/template"
	"time"

	qrcode "github.com/skip2/go-qrcode"
	"gopkg.in/mail.v2"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventService struct {
	EventRepo *repos.EventRepo
}

func NewEventService(repo *repos.EventRepo) *EventService {
	return &EventService{
		EventRepo: repo,
	}
}

func (s *EventService) CreateEvent(user models.User, body models.CreateEventRequest) (*models.Event, error) {
	if !user.IsEventCreator && !user.IsSuperUser {
		return nil, errors.New("only super users or event creators can create events")
	}

	var event models.Event
	event.ID = uuid.New().String()
	event.CreatedBy = user.ID

	if body.Slug == "" {
		return nil, errors.New("event slug can't be empty")
	}

	if body.EndDate.Before(body.StartDate) {
		return nil, errors.New("event end can't be before event start")
	}

	event.Name = body.Name
	event.Slug = strings.ToLower(body.Slug)
	event.Description = body.Description
	event.Location = body.Location
	event.StartDate = body.StartDate
	event.EndDate = body.EndDate
	event.IsPublic = true
	event.IsHidden = body.IsHidden
	event.IsBlocked = body.IsBlocked
	event.MaxTokensPerUser = body.MaxTokensPerUser

	err := s.EventRepo.CreateEvent(&event)
	return &event, err
}

func (s *EventService) GetEvent(slug string) (*models.Event, error) {
	return s.EventRepo.GetEventBySlug(slug)
}

func (s *EventService) GetAllEvents() ([]models.Event, error) {
	return s.EventRepo.GetAllEvents()
}

func (s *EventService) UpdateEvent(user models.User, slug string, newData *models.UpdateEventRequest) (*models.Event, error) {
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return nil, err
	}

	if !user.IsSuperUser {
		if event.CreatedBy != user.ID {
			return nil, errors.New("event can only be edited by its creator")
		}
	}

	if newData.Slug == "" {
		return nil, errors.New("event slug can't be empty")
	}

	if newData.EndDate.Before(newData.StartDate) {
		return nil, errors.New("event end can't be before event start")
	}

	event.Name = newData.Name
	event.Slug = strings.ToLower(newData.Slug)
	event.Description = newData.Description
	event.Location = newData.Location
	event.StartDate = newData.StartDate
	event.EndDate = newData.EndDate
	event.IsHidden = newData.IsHidden
	event.IsBlocked = newData.IsBlocked
	event.MaxTokensPerUser = newData.MaxTokensPerUser

	err = s.EventRepo.UpdateEvent(event)
	return event, err
}

func (s *EventService) DeleteEvent(user models.User, slug string) error {
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	if !user.IsSuperUser {
		if event.CreatedBy != user.ID {
			return errors.New("only the event creator can delete it")
		}
	}

	products, err := s.EventRepo.GetEventBoughtProductsIDs(event.ID)
	if err != nil {
		return err
	}

	if len(products) > 0 {
		return errors.New("event has products that were bought, cannot delete")
	}

	attencances, err := s.GetAllAttendances(user, slug)
	if err != nil {
		return err
	}

	if len(attencances) > 0 {
		return errors.New("cannot delete the event if it has activities that have been attended")
	}

	return s.EventRepo.DeleteEvent(slug)
}

func (s *EventService) RegisterUserToEvent(user models.User, slug string) error {
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	if event.IsBlocked {
		return errors.New("event is blocked and not accepting registrations")
	}

	isRegistered, err := s.EventRepo.IsUserRegisteredToEvent(user.ID, slug)
	if err != nil {
		return err
	}
	if isRegistered {
		return errors.New("user already registered to this event")
	}

	registration := models.EventRegistration{
		EventID:      event.ID,
		UserID:       user.ID,
		RegisteredAt: time.Now(),
	}

	go func() {
		if err := s.SendRegistrationEmail(&user, event); err != nil {
			fmt.Printf("Failed to send registration email: %v\n", err)
		}
	}()

	event.ParticipantCount++
	s.EventRepo.UpdateEvent(event)

	go func() {
		activities, err := s.EventRepo.GetAllActivitiesFromEvent(event.ID)
		if err != nil {
			fmt.Printf("Failed to get activities for event %s: %v\n", event.ID, err)
			return
		}

		for _, activity := range activities {
			if activity.IsMandatory {
				activityRegistration := models.ActivityRegistration{
					ActivityID:   activity.ID,
					UserID:       user.ID,
					RegisteredAt: time.Now(),
					AccessMethod: "event",
				}

				if err := s.EventRepo.RegisterUserToActivity(&activityRegistration); err != nil {
					fmt.Printf("Failed to register user %s to mandatory activity %s: %v\n",
						user.ID, activity.ID, err)
				}
			}
		}
	}()

	return s.EventRepo.CreateEventRegistration(&registration)
}

func (s *EventService) SendRegistrationEmail(user *models.User, event *models.Event) error {
	from := config.GetSystemEmail()
	password := config.GetSystemEmailPass()

	// Generate QR code as PNG
	var png []byte
	png, err := qrcode.Encode(user.ID, qrcode.Medium, 256)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %v", err)
	}

	// Create a unique filename using user's name and timestamp
	safeFirstName := strings.ReplaceAll(user.Name, " ", "_")
	safeLastName := strings.ReplaceAll(user.LastName, " ", "_")
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s_%s_%d", safeFirstName, safeLastName, timestamp)

	// Read the template
	templatePath := filepath.Join("templates", "registration_email.html")
	file, err := os.Open(templatePath)
	if err != nil {
		return fmt.Errorf("failed to open email template: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read email template: %v", err)
	}

	tmpl, err := template.New("emailTemplate").Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Prepare template data with filename for QR code
	data := struct {
		User     models.User
		Event    models.Event
		Filename string
	}{
		User:     *user,
		Event:    *event,
		Filename: filename,
	}

	var body strings.Builder
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Create email using gomail
	m := mail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", "Registration to "+event.Name)
	m.SetBody("text/html", body.String())

	// Embed the QR code image
	m.EmbedReader(filename, strings.NewReader(string(png)))

	// Create dialer
	d := mail.NewDialer("smtp.gmail.com", 587, from, password)
	d.StartTLSPolicy = mail.MandatoryStartTLS

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func (s *EventService) UnregisterUserFromEvent(user models.User, slug string) error {
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	isRegistered, err := s.EventRepo.IsUserRegisteredToEvent(user.ID, slug)
	if err != nil {
		return err
	}
	if !isRegistered {
		return errors.New("user is not registered to this event")
	}

	productsRelation, err := s.EventRepo.GetUserProductsRelation(user.ID)
	if err != nil {
		return err
	}
	products, err := s.EventRepo.GetProductsFromUserProducts(productsRelation)
	if err != nil {
		return err
	}
	if len(products) > 0 {
		return errors.New("cannot unregister from event where you bought products")
	}

	actvities, err := s.EventRepo.GetUserAttendedActivities(user.ID)
	if err != nil {
		return err
	}
	if len(actvities) > 0 {
		return errors.New("cannot unregister from event where you attended activities")
	}

	if event.ParticipantCount > 0 {
		event.ParticipantCount--
		s.EventRepo.UpdateEvent(event)
	}

	return s.EventRepo.DeleteEventRegistration(user.ID, event.ID)
}

func (s *EventService) IsUserRegisteredToEvent(user models.User, slug string) (bool, error) {
	return s.EventRepo.IsUserRegisteredToEvent(user.ID, slug)
}

func (s *EventService) IsAdminTypeOf(user models.User, adminType models.AdminType, slug string) (bool, error) {
	adminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(user.ID, slug)
	if err != nil {
		return false, err
	}

	return adminStatus.AdminType == adminType, nil
}

func (s *EventService) PromoteUserOfEventBySlug(requester models.User, email string, slug string) error {
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	targetUser, err := s.EventRepo.GetUserByEmail(email)
	if err != nil {
		return errors.New("target user not found: " + err.Error())
	}

	if targetUser.ID == requester.ID {
		return errors.New("users cannot promote themselves")
	}

	isCreator := event.CreatedBy == targetUser.ID
	if isCreator || targetUser.IsSuperUser {
		return errors.New("cannot promote event creator or super user")
	}

	isRegistered, err := s.EventRepo.IsUserRegisteredToEvent(targetUser.ID, slug)
	if err != nil {
		return err
	}
	if !isRegistered {
		return errors.New("user must be registered to the event to be promoted")
	}

	if requester.IsSuperUser || event.CreatedBy == requester.ID {
		adminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(targetUser.ID, slug)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		if adminStatus == nil {
			return s.EventRepo.MakeAdminOfEventBySlug(targetUser.ID, slug)
		}

		if adminStatus.AdminType == models.AdminTypeNormal {
			return s.EventRepo.PromoteUserOfEventBySlug(targetUser.ID, slug)
		}

		return errors.New("user is already a master admin")
	}

	isMasterAdmin, err := s.IsAdminTypeOf(requester, models.AdminTypeMaster, slug)
	if err != nil {
		return err
	}

	if isMasterAdmin {
		adminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(targetUser.ID, slug)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		if adminStatus != nil {
			return errors.New("master admins can only promote to normal admin, user already has admin status")
		}

		return s.EventRepo.MakeAdminOfEventBySlug(targetUser.ID, slug)
	}

	return errors.New("only super users, event creators, or master admins can promote users")
}

func (s *EventService) DemoteUserOfEventBySlug(requester models.User, email string, slug string) error {
	event, err := s.EventRepo.GetEventBySlug(slug)
	if err != nil {
		return err
	}

	targetUser, err := s.EventRepo.GetUserByEmail(email)
	if err != nil {
		return errors.New("target user not found: " + err.Error())
	}

	if targetUser.ID == requester.ID {
		return errors.New("users cannot demote themselves")
	}

	isCreator := event.CreatedBy == targetUser.ID
	if isCreator || targetUser.IsSuperUser {
		return errors.New("cannot demote event creator or super user")
	}

	adminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(targetUser.ID, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("user is not an admin of this event")
		}
		return err
	}

	if requester.IsSuperUser || event.CreatedBy == requester.ID {
		if adminStatus.AdminType == models.AdminTypeNormal {
			return s.EventRepo.RemoveAdminOfEventBySlug(targetUser.ID, slug)
		}

		return s.EventRepo.DemoteUserOfEventBySlug(targetUser.ID, slug)
	}

	isMasterAdmin, err := s.IsAdminTypeOf(requester, models.AdminTypeMaster, slug)
	if err != nil {
		return err
	}

	if isMasterAdmin {
		if adminStatus.AdminType == models.AdminTypeMaster {
			return errors.New("master admins cannot demote other master admins")
		}

		return s.EventRepo.RemoveAdminOfEventBySlug(targetUser.ID, slug)
	}

	return errors.New("only super users, event creators, or master admins can demote users")
}

func (s *EventService) GetAllPublicEvents() ([]models.Event, error) {
	return s.EventRepo.GetAllPublicEvents()
}

func (s *EventService) GetUserByID(userID string) (models.User, error) {
	return s.EventRepo.GetUserByID(userID)
}

func (s *EventService) GetEventsCreatedByUser(user models.User) ([]models.Event, error) {
	return s.EventRepo.GetEventsCreatedByUser(user.ID)
}

func (s *EventService) GetUserEvents(user models.User) ([]models.Event, error) {
	return s.EventRepo.GetUserEvents(user.ID)
}

func (s *EventService) GetAllAttendances(admin models.User, eventSlug string) ([]models.ActivityRegistration, error) {
	event, err := s.EventRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	// Check admin permissions
	if !admin.IsSuperUser && event.CreatedBy != admin.ID {
		adminStatus, err := s.EventRepo.GetUserAdminStatusBySlug(admin.ID, eventSlug)
		if err != nil || (adminStatus.AdminType != models.AdminTypeMaster && adminStatus.AdminType != models.AdminTypeNormal) {
			return nil, errors.New("unauthorized: only admins can get all attendances")
		}
	}

	attendances, err := s.EventRepo.GetAllAttendancesFromEvent(event.ID)
	if err != nil {
		return nil, errors.New("failed to retrieve all attendances: " + err.Error())
	}

	return attendances, nil
}
