import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Métricas personalizadas
const errorRate = new Rate('errors');
const requestDuration = new Trend('request_duration');

// Configuração do teste
export const options = {
  scenarios: {
    api_test: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: 1,
      maxDuration: '5m',
    },
  },
  thresholds: {
    'http_req_duration': ['p(95)<2000'],
    'errors': ['rate<0.1'],
    'request_duration': ['p(95)<2000'],
  },
};

// Configurações
const BASE_URL = 'http://localhost:8080';
const SLEEP_TIME = 1;

// Tipos de atividades disponíveis
const ACTIVITY_TYPES = {
  PALESTRA: 'palestra',
  MINICURSO: 'minicurso',
  WORKSHOP: 'workshop',
  MESA_REDONDA: 'mesa_redonda',
  VISITA_TECNICA: 'visita_tecnica',
  OUTRO: 'outro'
};

/**
 * Generates a unique string by combining a prefix, the current timestamp, and a random suffix.
 *
 * @param {string} prefix - The prefix to include in the generated string.
 * @returns {string} A unique string suitable for test data or identifiers.
 */
function generateUniqueData(prefix) {
  const timestamp = new Date().getTime();
  const random = randomString(5);
  return `${prefix}-${timestamp}-${random}`;
}

/**
 * Creates a JSON string payload for user registration with the provided details.
 *
 * @param {string} email - The user's email address.
 * @param {string} password - The user's password.
 * @param {string} name - The user's first name.
 * @param {string} lastName - The user's last name.
 * @returns {string} A JSON string representing the registration payload.
 */
function createRegisterPayload(email, password, name, lastName) {
  return JSON.stringify({
    email: email,
    password: password,
    name: name,
    last_name: lastName
  });
}

/**
 * Generates a JSON string payload for user login with the provided email and password.
 *
 * @param {string} email - The user's email address.
 * @param {string} password - The user's password.
 * @returns {string} A JSON string containing the login credentials.
 */
function createLoginPayload(email, password) {
  return JSON.stringify({
    email: email,
    password: password
  });
}

/**
 * Constructs HTTP headers for authenticated API requests using access and refresh tokens.
 *
 * @param {string} authToken - The access token for the Authorization header.
 * @param {string} refreshToken - The refresh token for the Refresh header.
 * @returns {Object} An object containing the required headers for authenticated requests.
 */
function createAuthHeaders(authToken, refreshToken) {
  return {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
    'Refresh': `Bearer ${refreshToken}`,
  };
}

/**
 * Updates the user's authentication and refresh tokens if new tokens are present in the response headers.
 *
 * @param {Object} user - The user object whose tokens may be updated.
 * @param {Object} response - The HTTP response containing potential new tokens in headers.
 */
function checkAndUpdateTokens(user, response) {
  const newAccessToken = response.headers['X-New-Access-Token'];
  const newRefreshToken = response.headers['X-New-Refresh-Token'];

  if (newAccessToken) {
    console.log(`[${user.role}] Atualizando access token`);
    user.authToken = newAccessToken;
  }
  if (newRefreshToken) {
    console.log(`[${user.role}] Atualizando refresh token`);
    user.refreshToken = newRefreshToken;
  }
}

/**
 * Registers a new user by sending a POST request to the API.
 *
 * @param {string} email - The user's email address.
 * @param {string} password - The user's password.
 * @param {string} name - The user's first name.
 * @param {string} lastName - The user's last name.
 * @returns {boolean} True if registration is successful (HTTP 201), otherwise false.
 */
function registerUser(email, password, name, lastName) {
  console.log(`\n[System] Registrando novo usuário: ${email}`);
  const payload = createRegisterPayload(email, password, name, lastName);
  const res = http.post(`${BASE_URL}/register`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'registro bem sucedido': (r) => r.status === 201,
  });

  if (res.status === 201) {
    console.log(`[System] Usuário ${email} registrado com sucesso`);
    return true;
  } else {
    console.log(`[System] Falha ao registrar usuário ${email}:`, res.body);
    return false;
  }
}

/**
 * Authenticates a user by sending login credentials and updates the user object with access and refresh tokens on success.
 *
 * @param {Object} user - The user object containing email, password, and role.
 * @returns {boolean} True if authentication succeeds and tokens are set; otherwise, false.
 */
function authenticateUser(user) {
  console.log(`\n[${user.role}] Tentando fazer login com: ${user.email}`);
  
  const loginPayload = createLoginPayload(user.email, user.password);
  const loginRes = http.post(`${BASE_URL}/login`, loginPayload, {
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
  });

  if (loginRes.status === 200) {
    try {
      const data = JSON.parse(loginRes.body);
      if (data.data && data.data.access_token && data.data.refresh_token) {
        console.log(`[${user.role}] Login bem sucedido`);
        user.authToken = data.data.access_token;
        user.refreshToken = data.data.refresh_token;
        return true;
      }
    } catch (e) {
      console.log(`[${user.role}] Erro ao fazer parse da resposta do login:`, e);
    }
  }
  
  console.log(`[${user.role}] Login falhou`);
  return false;
}

/**
 * Creates a new event using the provided user credentials and event data.
 *
 * Sends an authenticated POST request to the `/events` endpoint to create an event. Updates the user's authentication tokens if new ones are provided in the response.
 *
 * @param {Object} user - The user object containing authentication tokens and role.
 * @param {Object} eventData - The event details to be created.
 * @returns {Object|null} The created event data if successful, or `null` if creation fails.
 */
function createEvent(user, eventData) {
  console.log(`\n[${user.role}] Criando evento: ${eventData.name}`);
  
  const params = {
    headers: createAuthHeaders(user.authToken, user.refreshToken),
  };

  const res = http.post(`${BASE_URL}/events`, JSON.stringify(eventData), params);
  checkAndUpdateTokens(user, res);

  check(res, {
    'evento criado com sucesso': (r) => r.status === 201,
  });

  if (res.status === 201) {
    console.log(`[${user.role}] Evento ${eventData.name} criado com sucesso`);
    return JSON.parse(res.body).data;
  } else {
    console.log(`[${user.role}] Falha ao criar evento:`, res.body);
    return null;
  }
}

/**
 * Registers a user for a specified event using their authentication tokens.
 *
 * @param {Object} user - The user object containing authentication tokens and role.
 * @param {string} eventSlug - The unique identifier for the event.
 * @returns {boolean} True if registration is successful; otherwise, false.
 */
function registerToEvent(user, eventSlug) {
  console.log(`\n[${user.role}] Registrando no evento: ${eventSlug}`);
  
  const params = {
    headers: createAuthHeaders(user.authToken, user.refreshToken),
  };

  const res = http.post(`${BASE_URL}/events/${eventSlug}/register`, null, params);
  checkAndUpdateTokens(user, res);

  check(res, {
    'registro no evento bem sucedido': (r) => r.status === 200,
  });

  if (res.status === 200) {
    console.log(`[${user.role}] Registro no evento ${eventSlug} bem sucedido`);
    return true;
  } else {
    console.log(`[${user.role}] Falha ao registrar no evento:`, res.body);
    return false;
  }
}

/**
 * Creates a new activity within a specified event using the provided user credentials.
 *
 * Sends an authenticated POST request to add an activity to the event identified by {@link eventSlug}. Updates the user's authentication tokens if new ones are returned.
 *
 * @param {Object} user - The user object containing authentication tokens and role.
 * @param {string} eventSlug - The unique identifier for the event.
 * @param {Object} activityData - The activity details to be created.
 * @returns {Object|null} The created activity data if successful, or {@code null} if creation fails.
 */
function createActivity(user, eventSlug, activityData) {
  console.log(`\n[${user.role}] Criando atividade: ${activityData.name}`);
  
  const params = {
    headers: createAuthHeaders(user.authToken, user.refreshToken),
  };

  const res = http.post(`${BASE_URL}/events/${eventSlug}/activity`, JSON.stringify(activityData), params);
  checkAndUpdateTokens(user, res);

  check(res, {
    'atividade criada com sucesso': (r) => r.status === 201,
  });

  if (res.status === 201) {
    console.log(`[${user.role}] Atividade ${activityData.name} criada com sucesso`);
    return JSON.parse(res.body).data;
  } else {
    console.log(`[${user.role}] Falha ao criar atividade:`, res.body);
    return null;
  }
}

/**
 * Registers a user for a specific activity within an event.
 *
 * Attempts to register the given user for the specified activity in the event identified by {@link eventSlug}. Updates the user's authentication tokens if new ones are provided in the response.
 *
 * @param {Object} user - The user object containing authentication tokens and role.
 * @param {string} eventSlug - The unique identifier for the event.
 * @param {string} activityId - The unique identifier for the activity.
 * @returns {boolean} True if registration is successful (HTTP 200), otherwise false.
 */
function registerForActivity(user, eventSlug, activityId) {
  console.log(`\n[${user.role}] Registrando para atividade: ${activityId}`);
  
  const params = {
    headers: createAuthHeaders(user.authToken, user.refreshToken),
  };

  const payload = JSON.stringify({
    activity_id: activityId
  });

  const res = http.post(`${BASE_URL}/events/${eventSlug}/activity/register`, payload, params);
  checkAndUpdateTokens(user, res);

  check(res, {
    'registro em atividade bem sucedido': (r) => r.status === 200,
  });

  if (res.status === 200) {
    console.log(`[${user.role}] Registro na atividade ${activityId} bem sucedido`);
    return true;
  } else {
    console.log(`[${user.role}] Falha ao registrar na atividade:`, res.body);
    return false;
  }
}

/**
 * Attempts to promote a user within a specific event.
 *
 * Sends an authenticated request to elevate the role of the user identified by {@link targetEmail} in the event specified by {@link eventSlug}. Returns true if the promotion succeeds (HTTP 200), otherwise false.
 *
 * @param {Object} promoter - The user performing the promotion, containing authentication tokens and role.
 * @param {string} eventSlug - The unique identifier of the event.
 * @param {string} targetEmail - The email address of the user to be promoted.
 * @returns {boolean} True if the promotion is successful; otherwise, false.
 */
function promoteUserInEvent(promoter, eventSlug, targetEmail) {
  console.log(`\n[${promoter.role}] Promovendo usuário ${targetEmail} no evento ${eventSlug}`);
  
  const params = {
    headers: createAuthHeaders(promoter.authToken, promoter.refreshToken),
  };

  const payload = JSON.stringify({
    email: targetEmail
  });

  const res = http.post(`${BASE_URL}/events/${eventSlug}/promote`, payload, params);
  checkAndUpdateTokens(promoter, res);

  check(res, {
    'promoção bem sucedida': (r) => r.status === 200,
  });

  if (res.status === 200) {
    console.log(`[${promoter.role}] Usuário ${targetEmail} promovido com sucesso`);
    return true;
  } else {
    console.log(`[${promoter.role}] Falha ao promover usuário:`, res.body);
    return false;
  }
}

/**
 * Demotes a user within a specific event by sending an authenticated request.
 *
 * Attempts to lower the role of the user identified by {@link targetEmail} in the event specified by {@link eventSlug}, using the credentials of the demoter.
 *
 * @param {Object} demoter - The user performing the demotion, containing authentication tokens and role.
 * @param {string} eventSlug - The unique identifier of the event.
 * @param {string} targetEmail - The email address of the user to be demoted.
 * @returns {boolean} True if the demotion was successful (HTTP 200), otherwise false.
 */
function demoteUserInEvent(demoter, eventSlug, targetEmail) {
  console.log(`\n[${demoter.role}] Rebaixando usuário ${targetEmail} no evento ${eventSlug}`);
  
  const params = {
    headers: createAuthHeaders(demoter.authToken, demoter.refreshToken),
  };

  const payload = JSON.stringify({
    email: targetEmail
  });

  const res = http.post(`${BASE_URL}/events/${eventSlug}/demote`, payload, params);
  checkAndUpdateTokens(demoter, res);

  check(res, {
    'rebaixamento bem sucedido': (r) => r.status === 200,
  });

  if (res.status === 200) {
    console.log(`[${demoter.role}] Usuário ${targetEmail} rebaixado com sucesso`);
    return true;
  } else {
    console.log(`[${demoter.role}] Falha ao rebaixar usuário:`, res.body);
    return false;
  }
}

// Função principal de teste
export default function() {
  // Gerar dados únicos para este teste
  const testId = generateUniqueData('test');
  
  // Definir usuários de teste
  const users = [
    {
      role: 'super',
      email: 'sctiuenf@gmail.com',
      password: 'ExamplePass#01',
      name: 'Super',
      lastName: 'User',
      authToken: '',
      refreshToken: ''
    },
    {
      role: 'admin',
      email: `admin-${testId}@example.com`,
      password: 'AdminPass#01',
      name: 'Admin',
      lastName: 'User',
      authToken: '',
      refreshToken: ''
    },
    {
      role: 'creator',
      email: `creator-${testId}@example.com`,
      password: 'CreatorPass#01',
      name: 'Creator',
      lastName: 'User',
      authToken: '',
      refreshToken: ''
    },
    {
      role: 'user1',
      email: `user1-${testId}@example.com`,
      password: 'UserPass#01',
      name: 'Regular',
      lastName: 'User 1',
      authToken: '',
      refreshToken: ''
    },
    {
      role: 'user2',
      email: `user2-${testId}@example.com`,
      password: 'UserPass#02',
      name: 'Regular',
      lastName: 'User 2',
      authToken: '',
      refreshToken: ''
    }
  ];

  // Registrar novos usuários (exceto super)
  for (let i = 1; i < users.length; i++) {
    const user = users[i];
    registerUser(user.email, user.password, user.name, user.lastName);
    sleep(SLEEP_TIME);
  }

  // Autenticar todos os usuários
  for (const user of users) {
    authenticateUser(user);
    sleep(SLEEP_TIME);
  }

  // Super user cria um evento
  const superUser = users[0];
  const eventData = {
    slug: `event-${testId}`,
    name: `Test Event ${testId}`,
    description: `Test event description ${testId}`,
    location: 'Test Location',
    start_date: new Date(Date.now() + 86400000).toISOString(), // Tomorrow
    end_date: new Date(Date.now() + 172800000).toISOString(), // Day after tomorrow
    max_tokens_per_user: 1,
    is_hidden: false,
    is_blocked: false
  };

  const event = createEvent(superUser, eventData);
  if (!event) {
    console.log('[System] Falha ao criar evento, abortando teste');
    return;
  }
  sleep(SLEEP_TIME);

  // Todos os usuários se registram no evento
  for (const user of users) {
    registerToEvent(user, event.slug);
    sleep(SLEEP_TIME);
  }

  // Criar diferentes tipos de atividades
  const activities = [];

  // 1. Palestra normal
  const palestraData = {
    name: `Palestra ${testId}`,
    description: `Palestra description ${testId}`,
    speaker: 'Test Speaker',
    location: 'Test Location',
    type: ACTIVITY_TYPES.PALESTRA,
    start_time: new Date(Date.now() + 86400000).toISOString(),
    end_time: new Date(Date.now() + 90000000).toISOString(),
    has_unlimited_capacity: false,
    max_capacity: 30,
    is_mandatory: false,
    has_fee: false,
    is_standalone: false,
    is_hidden: false,
    is_blocked: false
  };
  activities.push(createActivity(superUser, event.slug, palestraData));
  sleep(SLEEP_TIME);

  // 2. Minicurso com taxa
  const minicursoData = {
    name: `Minicurso ${testId}`,
    description: `Minicurso description ${testId}`,
    speaker: 'Test Speaker',
    location: 'Test Location',
    type: ACTIVITY_TYPES.MINICURSO,
    start_time: new Date(Date.now() + 86400000).toISOString(),
    end_time: new Date(Date.now() + 90000000).toISOString(),
    has_unlimited_capacity: false,
    max_capacity: 20,
    is_mandatory: false,
    has_fee: true,
    is_standalone: false,
    is_hidden: false,
    is_blocked: false
  };
  activities.push(createActivity(superUser, event.slug, minicursoData));
  sleep(SLEEP_TIME);

  // 3. Workshop com capacidade limitada
  const workshopData = {
    name: `Workshop ${testId}`,
    description: `Workshop description ${testId}`,
    speaker: 'Test Speaker',
    location: 'Test Location',
    type: ACTIVITY_TYPES.WORKSHOP,
    start_time: new Date(Date.now() + 86400000).toISOString(),
    end_time: new Date(Date.now() + 90000000).toISOString(),
    has_unlimited_capacity: false,
    max_capacity: 10,
    is_mandatory: false,
    has_fee: false,
    is_standalone: false,
    is_hidden: false,
    is_blocked: false
  };
  activities.push(createActivity(superUser, event.slug, workshopData));
  sleep(SLEEP_TIME);

  // Testar promoções e rebaixamentos
  console.log('\n[System] Iniciando testes de promoção e rebaixamento');

  // 1. Super user promove admin para master admin
  promoteUserInEvent(superUser, event.slug, users[1].email);
  sleep(SLEEP_TIME);

  // 2. Master admin tenta promover creator (deve falhar)
  promoteUserInEvent(users[1], event.slug, users[2].email);
  sleep(SLEEP_TIME);

  // 3. Super user promove creator para admin
  promoteUserInEvent(superUser, event.slug, users[2].email);
  sleep(SLEEP_TIME);

  // 4. Admin tenta promover user1 (deve funcionar)
  promoteUserInEvent(users[2], event.slug, users[3].email);
  sleep(SLEEP_TIME);

  // 5. Admin tenta rebaixar master admin (deve falhar)
  demoteUserInEvent(users[2], event.slug, users[1].email);
  sleep(SLEEP_TIME);

  // 6. Super user rebaixa master admin
  demoteUserInEvent(superUser, event.slug, users[1].email);
  sleep(SLEEP_TIME);

  // Testar registros em atividades
  console.log('\n[System] Iniciando testes de registro em atividades');

  // 1. Usuários tentam se registrar na palestra (deve funcionar)
  for (const user of users) {
    registerForActivity(user, event.slug, activities[0].id);
    sleep(SLEEP_TIME);
  }

  // 2. Usuários tentam se registrar no minicurso pago (deve falhar sem tokens)
  for (const user of users) {
    registerForActivity(user, event.slug, activities[1].id);
    sleep(SLEEP_TIME);
  }

  // 3. Usuários tentam se registrar no workshop (deve falhar após atingir capacidade)
  for (const user of users) {
    registerForActivity(user, event.slug, activities[2].id);
    sleep(SLEEP_TIME);
  }
} 