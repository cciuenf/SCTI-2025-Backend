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

// Função para gerar dados únicos para cada teste
function generateUniqueData(prefix) {
  const timestamp = new Date().getTime();
  const random = randomString(5);
  return `${prefix}-${timestamp}-${random}`;
}

// Função para criar payload de registro
function createRegisterPayload(email, password, name, lastName) {
  return JSON.stringify({
    email: email,
    password: password,
    name: name,
    last_name: lastName
  });
}

// Função para criar payload de login
function createLoginPayload(email, password) {
  return JSON.stringify({
    email: email,
    password: password
  });
}

// Função para criar headers de autenticação
function createAuthHeaders(authToken, refreshToken) {
  return {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${authToken}`,
    'Refresh': `Bearer ${refreshToken}`,
  };
}

// Função para verificar e atualizar tokens
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

// Função para registrar um novo usuário
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

// Função para autenticar usuário
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

// Função para criar um evento
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

// Função para registrar em um evento
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

// Função para criar uma atividade
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

// Função para registrar em uma atividade
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

// Função para promover usuário em um evento
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

// Função para rebaixar usuário em um evento
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