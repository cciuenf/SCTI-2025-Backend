package services

import (
	"errors"
	"scti/internal/models"
	repos "scti/internal/repositories"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityService struct {
	ActivityRepo *repos.ActivityRepo
}

func NewActivityService(activityRepo *repos.ActivityRepo) *ActivityService {
	return &ActivityService{
		ActivityRepo: activityRepo,
	}
}

func (s *ActivityService) CreateEventActivity(user models.User, eventSlug string, req models.CreateActivityRequest) (*models.Activity, error) {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	if event.CreatedBy != user.ID && !user.IsSuperUser {
		isMasterAdmin, err := s.ActivityRepo.GetUserAdminStatusBySlug(user.ID, eventSlug)
		if err != nil || isMasterAdmin.AdminType != models.AdminTypeMaster {
			return nil, errors.New("unauthorized to create activities for this event")
		}
	}

	if req.EndTime.Before(req.StartTime) {
		return nil, errors.New("activity end time cannot be before start time")
	}

	if req.StartTime.Before(event.StartDate) || req.EndTime.After(event.EndDate) {
		return nil, errors.New("activity must be scheduled within event timeframe")
	}

	activity := models.Activity{
		ID:                   uuid.New().String(),
		EventID:              &event.ID,
		Name:                 req.Name,
		Description:          req.Description,
		Speaker:              req.Speaker,
		Location:             req.Location,
		Type:                 req.Type,
		StartTime:            req.StartTime,
		EndTime:              req.EndTime,
		HasUnlimitedCapacity: req.HasUnlimitedCapacity,
		MaxCapacity:          req.MaxCapacity,
		IsMandatory:          req.IsMandatory,
		HasFee:               req.HasFee,
		IsStandalone:         req.IsStandalone,
		IsHidden:             req.IsHidden,
		IsBlocked:            req.IsBlocked,
	}

	if req.IsStandalone {
		if req.StandaloneSlug == "" {
			return nil, errors.New("standalone activities must have a slug")
		}
		activity.StandaloneSlug = strings.ToLower(req.StandaloneSlug)
	}

	if err := s.ActivityRepo.CreateActivity(&activity); err != nil {
		return nil, errors.New("failed to create activity: " + err.Error())
	}

	return &activity, nil
}

func (s *ActivityService) GetAllActivitiesFromEvent(eventSlug string) ([]models.Activity, error) {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	activities, err := s.ActivityRepo.GetAllActivitiesFromEvent(event.ID)
	if err != nil {
		return nil, errors.New("failed to get activities: " + err.Error())
	}

	return activities, nil
}

func (s *ActivityService) UpdateEventActivity(user models.User, eventSlug string, activityID string, req models.ActivityUpdateRequest) (*models.Activity, error) {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	activity, err := s.ActivityRepo.GetActivityByID(activityID)
	if err != nil {
		return nil, errors.New("activity not found: " + err.Error())
	}

	if activity.EventID == nil || *activity.EventID != event.ID {
		return nil, errors.New("activity does not belong to this event")
	}

	if event.CreatedBy != user.ID && !user.IsSuperUser {
		isMasterAdmin, err := s.ActivityRepo.GetUserAdminStatusBySlug(user.ID, eventSlug)
		if err != nil || isMasterAdmin.AdminType != models.AdminTypeMaster {
			return nil, errors.New("unauthorized to update activities for this event")
		}
	}

	if req.EndTime.Before(req.StartTime) {
		return nil, errors.New("activity end time cannot be before start time")
	}

	if req.StartTime.Before(event.StartDate) || req.EndTime.After(event.EndDate) {
		return nil, errors.New("activity must be scheduled within event timeframe")
	}

	activity.Name = req.Name
	activity.Description = req.Description
	activity.Speaker = req.Speaker
	activity.Location = req.Location
	activity.Type = req.Type
	activity.StartTime = req.StartTime
	activity.EndTime = req.EndTime
	activity.HasUnlimitedCapacity = req.HasUnlimitedCapacity
	activity.MaxCapacity = req.MaxCapacity
	activity.IsMandatory = req.IsMandatory
	activity.HasFee = req.HasFee
	activity.IsHidden = req.IsHidden
	activity.IsBlocked = req.IsBlocked

	if req.IsStandalone {
		if req.StandaloneSlug == "" {
			return nil, errors.New("standalone activities must have a slug")
		}
		activity.IsStandalone = true
		activity.StandaloneSlug = strings.ToLower(req.StandaloneSlug)
	} else {
		activity.IsStandalone = false
		activity.StandaloneSlug = ""
	}

	if err := s.ActivityRepo.UpdateActivity(activity); err != nil {
		return nil, errors.New("failed to update activity: " + err.Error())
	}

	return activity, nil
}

func (s *ActivityService) DeleteEventActivity(user models.User, eventSlug string, activityID string) error {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return errors.New("event not found: " + err.Error())
	}

	activity, err := s.ActivityRepo.GetActivityByID(activityID)
	if err != nil {
		return errors.New("activity not found: " + err.Error())
	}

	if activity.EventID == nil || *activity.EventID != event.ID {
		return errors.New("activity does not belong to this event")
	}

	if event.CreatedBy != user.ID && !user.IsSuperUser {
		isMasterAdmin, err := s.ActivityRepo.GetUserAdminStatusBySlug(user.ID, eventSlug)
		if err != nil || isMasterAdmin.AdminType != models.AdminTypeMaster {
			return errors.New("unauthorized to delete activities for this event")
		}
	}

	registrations, err := s.ActivityRepo.GetActivityRegistrations(activityID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.New("failed to get activity registrations: " + err.Error())
	}

	if len(registrations) > 0 {
		return errors.New("activity has registrations")
	}

	if activity.StartTime.Before(time.Now()) {
		return errors.New("activity has already started")
	}

	if err := s.ActivityRepo.DeleteActivity(activityID); err != nil {
		return errors.New("failed to delete activity: " + err.Error())
	}

	return nil
}

func (s *ActivityService) RegisterUserToActivity(user models.User, eventSlug string, activityID string) error {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return errors.New("event not found: " + err.Error())
	}

	activity, err := s.ActivityRepo.GetActivityByID(activityID)
	if err != nil {
		return errors.New("activity not found: " + err.Error())
	}

	if activity.IsBlocked {
		return errors.New("activity is currently blocked")
	}

	if activity.EventID == nil || *activity.EventID != event.ID {
		return errors.New("activity does not belong to this event")
	}

	now := time.Now()
	if activity.EndTime.Before(now) {
		return errors.New("activity has already ended")
	}

	isRegistered, err := s.ActivityRepo.IsUserRegisteredToEvent(user.ID, event.Slug)
	if err != nil {
		return errors.New("error checking event registration: " + err.Error())
	}

	if !isRegistered {
		return errors.New("user must be registered to the event first")
	}

	if !activity.HasUnlimitedCapacity {
		currentRegistrations, maxCapacity, err := s.ActivityRepo.GetActivityCapacity(activityID)
		if err != nil {
			return errors.New("error checking activity capacity: " + err.Error())
		}

		if currentRegistrations >= maxCapacity {
			return errors.New("activity has reached maximum capacity")
		}
	}

	userActivities, err := s.GetUserActivities(user)
	if err != nil {
		return errors.New("couldn't get user activities")
	}
	for _, uAct := range userActivities {
		if !(uAct.EndTime.Before(activity.StartTime) || uAct.StartTime.After(activity.EndTime)) && uAct.Type != models.ActivityPalestra {
			return errors.New("user has another activity registered at the same time that is not palestra")
		}
	}

	userAccesses, err := s.ActivityRepo.GetUserAccesses(user.ID)
	if err != nil {
		return errors.New("error checking user accesses: " + err.Error())
	}

	var hasAccess bool
	for _, access := range userAccesses {
		if access.TargetID == activityID {
			hasAccess = true
			break
		}
	}

	if !hasAccess && activity.HasFee {
		userTokens, err := s.ActivityRepo.GetUserTokens(user.ID)
		if err != nil {
			return errors.New("error checking user tokens: " + err.Error())
		}

		if len(userTokens) == 0 {
			return errors.New("this activity requires a token or payment")
		}

		var useToken models.UserToken
		var foundToken bool
		for _, token := range userTokens {
			if !token.IsUsed && token.EventID == event.ID {
				useToken = token
				foundToken = true
				break
			}
		}

		if !foundToken {
			return errors.New("user does not have any available tokens")
		}

		useToken.IsUsed = true
		now := time.Now()
		useToken.UsedAt = &now
		useToken.UsedForID = &activityID
		if err := s.ActivityRepo.UpdateUserToken(useToken); err != nil {
			return errors.New("error updating user token: " + err.Error())
		}
	}

	registration := &models.ActivityRegistration{
		ActivityID:   activityID,
		UserID:       user.ID,
		AccessMethod: string(models.AccessMethodEvent), // Registered through event registration
	}

	if err := s.ActivityRepo.RegisterUserToActivity(registration); err != nil {
		return errors.New("failed to register to activity: " + err.Error())
	}

	return nil
}

func (s *ActivityService) UnregisterUserFromActivity(user models.User, eventSlug string, activityID string) error {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return errors.New("event not found: " + err.Error())
	}

	activity, err := s.ActivityRepo.GetActivityByID(activityID)
	if err != nil {
		return errors.New("activity not found: " + err.Error())
	}

	if activity.IsBlocked {
		return errors.New("activity is currently blocked")
	}

	if activity.EventID == nil || *activity.EventID != event.ID {
		return errors.New("activity does not belong to this event")
	}

	isRegistered, registration, err := s.ActivityRepo.IsUserRegisteredToActivity(activityID, user.ID)
	if err != nil {
		return errors.New("error checking activity registration: " + err.Error())
	}

	if !isRegistered {
		return errors.New("user is not registered to this activity")
	}

	if registration.AttendedAt != nil {
		return errors.New("user has already attended this activity")
	}

	userAccesses, err := s.ActivityRepo.GetUserAccesses(user.ID)
	if err != nil {
		return errors.New("error checking user accesses: " + err.Error())
	}

	var hasAccess bool
	for _, access := range userAccesses {
		if access.TargetID == activityID {
			hasAccess = true
			break
		}
	}

	if hasAccess {
		return errors.New("user has direct paid access to this activity")
	}

	if activity.HasFee {
		userTokens, err := s.ActivityRepo.GetUserTokens(user.ID)
		if err != nil {
			return errors.New("error checking user tokens: " + err.Error())
		}

		if len(userTokens) == 0 {
			return errors.New("this activity requires a token or payment")
		}

		var cleanToken models.UserToken
		var foundToken bool
		for _, token := range userTokens {
			if token.IsUsed {
				if *token.UsedForID == activityID {
					cleanToken = token
					foundToken = true
					break
				}
			}
		}

		if !foundToken {
			return errors.New("user does not have any available tokens")
		}

		cleanToken.IsUsed = false
		cleanToken.UsedAt = nil
		cleanToken.UsedForID = nil
		if err := s.ActivityRepo.UpdateUserToken(cleanToken); err != nil {
			return errors.New("error updating user token: " + err.Error())
		}
	}

	if err := s.ActivityRepo.UnregisterUserFromActivity(activityID, user.ID); err != nil {
		return errors.New("failed to unregister from activity: " + err.Error())
	}

	return nil
}

// TODO: When user buys the ticket to a standalone activity, we need to register them to the activity automatically
// so when that logic is implemented, we can remove the need to call this function and delete it
func (s *ActivityService) RegisterUserToStandaloneActivity(user models.User, eventSlug string, activityID string) error {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return errors.New("event not found: " + err.Error())
	}

	activity, err := s.ActivityRepo.GetActivityByID(activityID)
	if err != nil {
		return errors.New("activity not found: " + err.Error())
	}

	if activity.IsBlocked {
		return errors.New("activity is currently blocked")
	}

	if activity.EventID == nil || *activity.EventID != event.ID {
		return errors.New("activity does not belong to this event")
	}

	if !activity.IsStandalone {
		return errors.New("this activity does not support standalone registration")
	}

	if !activity.HasUnlimitedCapacity {
		currentRegistrations, maxCapacity, err := s.ActivityRepo.GetActivityCapacity(activityID)
		if err != nil {
			return errors.New("error checking activity capacity: " + err.Error())
		}

		if currentRegistrations >= maxCapacity {
			return errors.New("activity has reached maximum capacity")
		}
	}

	now := time.Now()
	if activity.EndTime.Before(now) {
		return errors.New("activity has already ended")
	}

	registration := &models.ActivityRegistration{
		ActivityID:               activityID,
		UserID:                   user.ID,
		IsStandaloneRegistration: true,
		AccessMethod:             string(models.AccessMethodDirect), // Direct registration without event registration
	}

	if err := s.ActivityRepo.RegisterUserToActivity(registration); err != nil {
		return errors.New("failed to register to activity: " + err.Error())
	}

	return nil
}

func (s *ActivityService) UnregisterUserFromStandaloneActivity(user models.User, eventSlug string, activityID string) error {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return errors.New("event not found: " + err.Error())
	}

	activity, err := s.ActivityRepo.GetActivityByID(activityID)
	if err != nil {
		return errors.New("activity not found: " + err.Error())
	}

	if activity.EventID == nil || *activity.EventID != event.ID {
		return errors.New("activity does not belong to this event")
	}

	if !activity.IsStandalone {
		return errors.New("this activity does not support standalone registration")
	}

	isRegistered, registration, err := s.ActivityRepo.IsUserRegisteredToActivity(activityID, user.ID)
	if err != nil {
		return errors.New("error checking activity registration: " + err.Error())
	}

	if !isRegistered {
		return errors.New("user is not registered to this activity")
	}

	if registration.AttendedAt != nil {
		return errors.New("user has already attended this activity")
	}

	userAccesses, err := s.ActivityRepo.GetUserAccesses(user.ID)
	if err != nil {
		return errors.New("error checking user accesses: " + err.Error())
	}

	var hasAccess bool
	for _, access := range userAccesses {
		if access.TargetID == activityID {
			hasAccess = true
			break
		}
	}

	if hasAccess {
		return errors.New("user has direct paid access to this activity")
	}

	if activity.IsBlocked {
		return errors.New("activity is currently blocked")
	}

	if err := s.ActivityRepo.UnregisterUserFromActivity(activityID, user.ID); err != nil {
		return errors.New("failed to unregister from activity: " + err.Error())
	}

	return nil
}

func (s *ActivityService) AttendActivity(admin models.User, eventSlug string, activityID string, userID string) error {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return errors.New("event not found: " + err.Error())
	}

	activity, err := s.ActivityRepo.GetActivityByID(activityID)
	if err != nil {
		return errors.New("activity not found: " + err.Error())
	}

	if activity.EventID == nil || *activity.EventID != event.ID {
		return errors.New("activity does not belong to this event")
	}

	if !admin.IsSuperUser && event.CreatedBy != admin.ID {
		adminStatus, err := s.ActivityRepo.GetUserAdminStatusBySlug(admin.ID, eventSlug)
		if err != nil || (adminStatus.AdminType != models.AdminTypeMaster && adminStatus.AdminType != models.AdminTypeNormal) {
			return errors.New("unauthorized: only admins can mark attendance")
		}
	}

	isRegistered, registration, err := s.ActivityRepo.IsUserRegisteredToActivity(activityID, userID)
	if err != nil {
		return errors.New("error checking activity registration: " + err.Error())
	}

	if !isRegistered {
		return errors.New("user is not registered to this activity")
	}

	if registration.AttendedAt != nil {
		return errors.New("user has already attended this activity")
	}

	if err := s.ActivityRepo.SetUserAttendance(activityID, userID, true); err != nil {
		return errors.New("failed to mark attendance: " + err.Error())
	}

	return nil
}

func (s *ActivityService) UnattendActivity(admin models.User, eventSlug string, activityID string, userID string) error {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return errors.New("event not found: " + err.Error())
	}

	activity, err := s.ActivityRepo.GetActivityByID(activityID)
	if err != nil {
		return errors.New("activity not found: " + err.Error())
	}

	if activity.EventID == nil || *activity.EventID != event.ID {
		return errors.New("activity does not belong to this event")
	}

	if !admin.IsSuperUser && event.CreatedBy != admin.ID {
		adminStatus, err := s.ActivityRepo.GetUserAdminStatusBySlug(admin.ID, eventSlug)
		if err != nil || adminStatus.AdminType != models.AdminTypeMaster {
			return errors.New("unauthorized: only master admins, event creators, or super users can remove attendance")
		}
	}

	isRegistered, registration, err := s.ActivityRepo.IsUserRegisteredToActivity(activityID, userID)
	if err != nil {
		return errors.New("error checking activity registration: " + err.Error())
	}

	if !isRegistered {
		return errors.New("user is not registered to this activity")
	}

	if registration.AttendedAt == nil {
		return errors.New("user has not attended this activity")
	}

	if err := s.ActivityRepo.SetUserAttendance(activityID, userID, false); err != nil {
		return errors.New("failed to remove attendance: " + err.Error())
	}

	return nil
}

func (s *ActivityService) GetActivityRegistrations(admin models.User, eventSlug string, activityID string) ([]models.ActivityRegistration, error) {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	activity, err := s.ActivityRepo.GetActivityByID(activityID)
	if err != nil {
		return nil, errors.New("activity not found: " + err.Error())
	}

	if activity.EventID == nil || *activity.EventID != event.ID {
		return nil, errors.New("activity does not belong to this event")
	}

	if !admin.IsSuperUser && event.CreatedBy != admin.ID {
		adminStatus, err := s.ActivityRepo.GetUserAdminStatusBySlug(admin.ID, eventSlug)
		if err != nil || (adminStatus.AdminType != models.AdminTypeMaster && adminStatus.AdminType != models.AdminTypeNormal) {
			return nil, errors.New("unauthorized: only admins can get activity attendees")
		}
	}

	var registrations []models.ActivityRegistration
	if registrations, err = s.ActivityRepo.GetActivityRegistrations(activityID); err != nil {
		return nil, errors.New("failed to retrieve activity registrations: " + err.Error())
	}

	return registrations, nil
}

func (s *ActivityService) GetUserAccesses(userID string) ([]models.AccessTarget, error) {
	return s.ActivityRepo.GetUserAccesses(userID)
}

func (s *ActivityService) GetUserAccessesFromEvent(userID string, eventSlug string) ([]models.AccessTarget, error) {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	return s.ActivityRepo.GetUserAccessesFromEvent(userID, *event)
}

func (s *ActivityService) GetUserActivities(user models.User) ([]models.Activity, error) {
	userActivities, err := s.ActivityRepo.GetUserActivities(user.ID)
	if err != nil {
		return nil, errors.New("error checking user activities: " + err.Error())
	}

	return userActivities, nil
}

func (s *ActivityService) GetUserActivitiesFromEvent(user models.User, eventSlug string) ([]models.Activity, error) {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	userActivities, err := s.ActivityRepo.GetUserActivities(user.ID)
	if err != nil {
		return nil, errors.New("error checking user activities: " + err.Error())
	}

	var activities []models.Activity
	for _, activity := range userActivities {
		if activity.EventID == &event.ID {
			activities = append(activities, activity)
		}
	}

	return activities, nil
}

func (s *ActivityService) GetActivityAttendants(admin models.User, eventSlug string, activityID string) ([]models.ActivityRegistration, error) {
	event, err := s.ActivityRepo.GetEventBySlug(eventSlug)
	if err != nil {
		return nil, errors.New("event not found: " + err.Error())
	}

	activity, err := s.ActivityRepo.GetActivityByID(activityID)
	if err != nil {
		return nil, errors.New("activity not found: " + err.Error())
	}

	if activity.EventID == nil || *activity.EventID != event.ID {
		return nil, errors.New("activity does not belong to this event")
	}

	if !admin.IsSuperUser && event.CreatedBy != admin.ID {
		adminStatus, err := s.ActivityRepo.GetUserAdminStatusBySlug(admin.ID, eventSlug)
		if err != nil || (adminStatus.AdminType != models.AdminTypeMaster && adminStatus.AdminType != models.AdminTypeNormal) {
			return nil, errors.New("unauthorized: only admins can get activity attendants")
		}
	}

	registrations, err := s.ActivityRepo.GetActivityRegistrations(activityID)
	if err != nil {
		return nil, errors.New("failed to retrieve activity attendants: " + err.Error())
	}

	var attendants []models.ActivityRegistration
	for _, registration := range registrations {
		if registration.AttendedAt != nil {
			attendants = append(attendants, registration)
		}
	}

	return attendants, nil
}
