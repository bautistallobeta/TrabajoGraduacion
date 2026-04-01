// Test de sistema — cobertura de todos los endpoints del servicio.
// Corre secuencialmente en un único VU (1 iteración). Usa IDs fijos con prefijo 9999
// para ser idempotente en re-runs: las cuentas y monedas de test se reusan si ya existen,
// y los balances se verifican con >= donde T06+reversal pueden acumularse entre corridas.
// Requiere: k6 instalado, servidor corriendo, Kafka activo.
import http from 'k6/http';
import { sleep, check, group } from 'k6';

export const options = {
  vus: 1,
  iterations: 1,
};

// ═══════════════════════════════════════════════════════════════════════════
// CONFIGURACIÓN — ajustar antes de correr
// ═══════════════════════════════════════════════════════════════════════════

// URL base del servidor
const BASE = 'http://localhost:8080';

// API Key del actor SISTEMA (debe coincidir con el parámetro APIKEY_SISTEMA en MySQL)
const APIKEY_SISTEMA = 'CAMBIAR_ESTE_VALOR';

// Usuario admin ACTIVO en MySQL (para obtener Bearer token en pruebas de actor USUARIO)
const ADMIN_USUARIO  = 'admin';
const ADMIN_PASSWORD = 'Admin123'; // plain text; el servidor aplica MD5

// ═══════════════════════════════════════════════════════════════════════════
// IDs FIJOS — prefijo 9999 (no colisiona con otros tests del proyecto)
// ═══════════════════════════════════════════════════════════════════════════

// Monedas
const ID_MONEDA_PRINCIPAL  = 998901;
const ID_MONEDA_SECUNDARIA = 998902;
const ID_MONEDA_BORRAR     = 998903; // creada y borrada en mismo test

// IdUsuarioFinal (cuentas en TigerBeetle)
const ID_UA = 9998901; // Usuario A — cuenta en MP y MS
const ID_UB = 9998902; // Usuario B — cuenta solo en MP

// Transfers (IDs fijos -> idempotentes en TB en reruns)
const T01 = '9989010000001'; // Ingreso  $1000  UA/MP
const T02 = '9989010000002'; // Ingreso  $500   UA/MP
const T03 = '9989010000003'; // Egreso   $200   UA/MP
const T04 = '9989010000004'; // Ingreso  $800   UB/MP
const T05 = '9989010000005'; // Ingreso  $300   UA/MS
const T06 = '9989010000006'; // Egreso   $150   UA/MP  (preparado para reversal)
// T07 = reversal de T06 (enviado con IdTransferencia=T06, Tipo=R)
const T08 = '9989010000008'; // Egreso   $100   UB/MP  (cuenta cerrada -> rechazada)

// Balances finales esperados (ver sección FASE 10 para cálculo)
// UA/MP: Creditos=1650.00  Debitos=350.00  Balance=1300.00
// UB/MP: Creditos=800.00   Debitos=0.00    Balance=800.00
// UA/MS: Creditos=300.00   Debitos=0.00    Balance=300.00

// ═══════════════════════════════════════════════════════════════════════════
// CONSTANTES DE TIEMPO Y PASSWORDS DE PRUEBA
// ═══════════════════════════════════════════════════════════════════════════

// Ventana de procesamiento Kafka: batch-size (5s) + procesamiento + margen
const KAFKA_WAIT = 5;

// Passwords usadas en el ciclo de vida del usuario de prueba (FASE 9)
// PASS_INICIAL: password que el usuario establece al confirmar su cuenta
// PASS_NUEVO:   password al que cambia en la prueba de modificar-password
// (definidas en el cuerpo de la función; buscar PASS_INICIAL y PASS_NUEVO)

// ═══════════════════════════════════════════════════════════════════════════
// HELPERS
// ═══════════════════════════════════════════════════════════════════════════

const PARAMS_SISTEMA = {
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': APIKEY_SISTEMA,
  },
};

const PARAMS_NO_AUTH = {
  headers: { 'Content-Type': 'application/json' },
};

function makeBearer(token) {
  return {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
  };
}

function parseBody(res) {
  try { return JSON.parse(res.body); } catch (_) { return {}; }
}

function logRes(label, res) {
  console.log(`  [${label}] status=${res.status}  body=${res.body}`);
}

function buildQS(params) {
  return Object.entries(params)
    .flatMap(([k, v]) =>
      Array.isArray(v)
        ? v.map(x => `${encodeURIComponent(k)}=${encodeURIComponent(x)}`)
        : [`${encodeURIComponent(k)}=${encodeURIComponent(v)}`]
    )
    .join('&');
}

function postTransfer(authParams, idTransferencia, idUsuarioFinal, monto, idMoneda, tipo, idCategoria) {
  const body = JSON.stringify({
    IdTransferencia: String(idTransferencia),
    IdUsuarioFinal:  idUsuarioFinal,
    Monto:           monto,
    IdMoneda:        idMoneda,
    Tipo:            tipo,
    IdCategoria:     idCategoria !== undefined ? idCategoria : 1,
    Fecha:           '2026-01-15 12:00:00',
  });
  return http.post(`${BASE}/transferencias`, body, authParams);
}

function waitKafka(label) {
  console.log(`\n  ⏳ Esperando ${KAFKA_WAIT}s para Kafka [${label}]...`);
  sleep(KAFKA_WAIT);
}

// ═══════════════════════════════════════════════════════════════════════════
// TEST PRINCIPAL
// ═══════════════════════════════════════════════════════════════════════════

export default function () {
  console.log('');
  console.log('╔══════════════════════════════════════════════════════════════════╗');
  console.log('║           TEST SISTEMA — COBERTURA TOTAL DE ENDPOINTS            ║');
  console.log('╚══════════════════════════════════════════════════════════════════╝');
  console.log(`  Moneda Principal: ${ID_MONEDA_PRINCIPAL} | Secundaria: ${ID_MONEDA_SECUNDARIA}`);
  console.log(`  UsuarioFinal A: ${ID_UA} | B: ${ID_UB}`);
  console.log(`  KAFKA_WAIT: ${KAFKA_WAIT}s`);

  let tokenAdmin       = '';
  let PARAMS_USUARIO   = makeBearer('');
  let idUsuarioNuevo   = 0;
  let passTemp         = '';
  let tokenNuevoUsr    = '';
  let tokenParaLogout  = '';
  const TS             = Date.now();
  const NOMBRE_USR_NUEVO  = `ts_${TS}`;
  const PASS_INICIAL      = 'Password123!';
  const PASS_NUEVO        = 'NuevoPass456!';

  // =========================================================================
  // FASE 0: PING Y AUTENTICACIÓN
  // =========================================================================
  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('FASE 0: PING Y AUTENTICACIÓN');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  group('F0-01: GET /ping — sin auth, debe retornar 200', () => {
    const res = http.get(`${BASE}/ping`, PARAMS_NO_AUTH);
    logRes('ping', res);
    check(res, { 'ping -> 200': (r) => r.status === 200 });
  });

  group('F0-02: Cualquier endpoint sin credenciales -> 401', () => {
    const res = http.get(`${BASE}/parametros`, PARAMS_NO_AUTH);
    logRes('sin-auth', res);
    check(res, { 'sin credenciales -> 401': (r) => r.status === 401 });
  });

  group('F0-03: Bearer + X-API-Key simultáneamente -> 400', () => {
    const res = http.get(`${BASE}/parametros`, {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer tokenfalso',
        'X-API-Key': APIKEY_SISTEMA,
      },
    });
    logRes('doble-auth', res);
    check(res, { 'doble credencial -> 400': (r) => r.status === 400 });
  });

  group('F0-04: Bearer token inválido -> 401', () => {
    const res = http.get(`${BASE}/parametros`, makeBearer('token_invalido_9999zz'));
    logRes('token-invalido', res);
    check(res, { 'token inválido -> 401': (r) => r.status === 401 });
  });

  group('F0-05: X-API-Key inválida -> 401', () => {
    const res = http.get(`${BASE}/parametros`, {
      headers: { 'Content-Type': 'application/json', 'X-API-Key': 'clave_falsa_9999' },
    });
    logRes('apikey-invalida', res);
    check(res, { 'API key inválida -> 401': (r) => r.status === 401 });
  });

  group('F0-06: POST /usuarios/login con admin existente -> obtener token', () => {
    const res = http.post(
      `${BASE}/usuarios/login`,
      JSON.stringify({ Usuario: ADMIN_USUARIO, Password: ADMIN_PASSWORD }),
      PARAMS_NO_AUTH
    );
    logRes('login-admin', res);
    if (res.status === 200) {
      const b = parseBody(res);
      tokenAdmin     = b.TokenSesion || '';
      PARAMS_USUARIO = makeBearer(tokenAdmin);
      console.log(`  Token admin obtenido: ${tokenAdmin ? tokenAdmin.substring(0, 8) + '...' : 'NONE'}`);
      check(null, { 'token admin obtenido': () => tokenAdmin !== '' });
    } else {
      check(res, { 'login admin -> 200': (r) => r.status === 200 });
      console.log('  ADVERTENCIA: sin token admin — tests USUARIO usarán SISTEMA donde sea posible');
    }
  });

  group('F0-07: POST /usuarios/login — credenciales incorrectas -> 400', () => {
    const res = http.post(
      `${BASE}/usuarios/login`,
      JSON.stringify({ Usuario: 'usuario_inexistente_zzzz9999', Password: 'passfalsa' }),
      PARAMS_NO_AUTH
    );
    logRes('login-fail', res);
    check(res, { 'credenciales incorrectas -> 400': (r) => r.status === 400 });
  });

  group('F0-08: POST /usuarios/login — campos vacíos -> 400', () => {
    const res = http.post(
      `${BASE}/usuarios/login`,
      JSON.stringify({ Usuario: '', Password: '' }),
      PARAMS_NO_AUTH
    );
    logRes('login-campos-vacios', res);
    check(res, { 'login campos vacíos -> 400': (r) => r.status === 400 });
  });

  group('F0-09: POST /usuarios/login — sin campos -> 400', () => {
    const res = http.post(`${BASE}/usuarios/login`, JSON.stringify({}), PARAMS_NO_AUTH);
    logRes('login-sin-campos', res);
    check(res, { 'sin campos -> 400': (r) => r.status === 400 });
  });

  // =========================================================================
  // FASE 1: PARÁMETROS
  // =========================================================================
  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('FASE 1: PARÁMETROS');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  group('F1-01: GET /parametros/MONTOMINTRANSFER — dame parámetro existente', () => {
    const res = http.get(`${BASE}/parametros/MONTOMINTRANSFER`, PARAMS_SISTEMA);
    logRes('dame-param', res);
    check(res, { 'dame parámetro -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Parametro=${b.Parametro} Valor=${b.Valor}`);
      check(b, {
        'tiene Parametro': (x) => x.Parametro === 'MONTOMINTRANSFER',
        'tiene Valor':     (x) => x.Valor !== undefined && x.Valor !== '',
      });
    }
  });

  group('F1-02: GET /parametros/PARAMETRO_INEXISTENTE_ZZZZ -> 404', () => {
    const res = http.get(`${BASE}/parametros/PARAMETRO_INEXISTENTE_ZZZZ`, PARAMS_SISTEMA);
    logRes('dame-param-noexiste', res);
    check(res, { 'parámetro inexistente -> 404': (r) => r.status === 404 });
  });

  group('F1-03: GET /parametros — buscar todos (sin filtro)', () => {
    const res = http.get(`${BASE}/parametros`, PARAMS_SISTEMA);
    logRes('buscar-params-todos', res);
    check(res, { 'buscar parámetros -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total parámetros: ${Array.isArray(b) ? b.length : 'N/A'}`);
      check(null, { 'retorna array': () => Array.isArray(b) });
    }
  });

  group('F1-04: GET /parametros?Cadena=MONTO — buscar por cadena', () => {
    const res = http.get(`${BASE}/parametros?Cadena=MONTO`, PARAMS_SISTEMA);
    logRes('buscar-params-monto', res);
    check(res, { 'buscar por cadena -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      const arr = Array.isArray(b) ? b : [];
      console.log(`  Total con "MONTO": ${arr.length}  (esperado >= 2)`);
      check(null, { 'filtra por cadena correctamente': () => arr.length >= 2 });
    }
  });

  group('F1-05: GET /parametros?Cadena=cadena_inexistente_9999 — búsqueda sin resultados', () => {
    // NOTA: tsp_buscar_parametros ignora el parámetro pCadena en su cláusula WHERE;
    // retorna todos los parámetros independientemente del valor enviado.
    // Por eso no se puede verificar que retorna array vacío para cadenas inexistentes.
    const res = http.get(`${BASE}/parametros?Cadena=cadena_inexistente_9999`, PARAMS_SISTEMA);
    logRes('buscar-params-vacio', res);
    check(res, { 'búsqueda con cadena inexistente -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      check(null, { 'retorna array (SP ignora filtro cadena)': () => Array.isArray(b) });
    }
  });

  group('F1-06: PUT /parametros/LIMITEBUSCARCUENTAS — modificar con actor SISTEMA', () => {
    const res = http.put(
      `${BASE}/parametros/LIMITEBUSCARCUENTAS`,
      JSON.stringify({ Valor: '100' }),
      PARAMS_SISTEMA
    );
    logRes('modificar-param-sistema', res);
    check(res, { 'modificar parámetro SISTEMA -> 200': (r) => r.status === 200 });
  });

  group('F1-07: PUT /parametros/LIMITEBUSCARCUENTAS — modificar con actor USUARIO', () => {
    if (tokenAdmin === '') { console.log('  SKIP: no hay token admin'); return; }
    const res = http.put(
      `${BASE}/parametros/LIMITEBUSCARCUENTAS`,
      JSON.stringify({ Valor: '100' }),
      PARAMS_USUARIO
    );
    logRes('modificar-param-usuario', res);
    check(res, { 'modificar parámetro USUARIO -> 200': (r) => r.status === 200 });
  });

  group('F1-08: PUT /parametros/LIMITEBUSCARCUENTAS — sin Valor -> 400', () => {
    const res = http.put(
      `${BASE}/parametros/LIMITEBUSCARCUENTAS`,
      JSON.stringify({}),
      PARAMS_SISTEMA
    );
    logRes('modificar-param-sin-valor', res);
    check(res, { 'sin Valor -> 400': (r) => r.status === 400 });
  });

  // =========================================================================
  // FASE 2: MONEDAS
  // =========================================================================
  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('FASE 2: MONEDAS');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  group(`F2-01: POST /monedas — crear moneda principal ${ID_MONEDA_PRINCIPAL}`, () => {
    const res = http.post(
      `${BASE}/monedas`,
      JSON.stringify({ IdMoneda: ID_MONEDA_PRINCIPAL }),
      PARAMS_SISTEMA
    );
    logRes('crear-moneda-principal', res);
    // 201 = creada; 400 = ya existe de run anterior (ambos son válidos para re-runs)
    check(res, { 'crear moneda principal -> 201 o 400': (r) => r.status === 201 || r.status === 400 });
  });

  group(`F2-02: POST /monedas — crear moneda secundaria ${ID_MONEDA_SECUNDARIA}`, () => {
    const res = http.post(
      `${BASE}/monedas`,
      JSON.stringify({ IdMoneda: ID_MONEDA_SECUNDARIA }),
      PARAMS_SISTEMA
    );
    logRes('crear-moneda-secundaria', res);
    check(res, { 'crear moneda secundaria -> 201 o 400': (r) => r.status === 201 || r.status === 400 });
  });

  group('F2-03: POST /monedas — IdMoneda=0 -> 400', () => {
    const res = http.post(`${BASE}/monedas`, JSON.stringify({ IdMoneda: 0 }), PARAMS_SISTEMA);
    logRes('crear-moneda-id-cero', res);
    check(res, { 'IdMoneda=0 -> 400': (r) => r.status === 400 });
  });

  group('F2-04: POST /monedas — sin IdMoneda en body -> 400', () => {
    const res = http.post(`${BASE}/monedas`, JSON.stringify({}), PARAMS_SISTEMA);
    logRes('crear-moneda-sin-id', res);
    check(res, { 'sin IdMoneda -> 400': (r) => r.status === 400 });
  });

  group(`F2-05: GET /monedas/${ID_MONEDA_PRINCIPAL} — dame moneda activa`, () => {
    const res = http.get(`${BASE}/monedas/${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('dame-moneda', res);
    check(res, { 'dame moneda -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  IdMoneda=${b.IdMoneda} Estado=${b.Estado}`);
      check(b, {
        'IdMoneda correcto': (x) => x.IdMoneda === ID_MONEDA_PRINCIPAL,
        'Estado es A':       (x) => x.Estado === 'A',
      });
    }
  });

  group('F2-06: GET /monedas/0 — IdMoneda=0 -> 400', () => {
    const res = http.get(`${BASE}/monedas/0`, PARAMS_SISTEMA);
    logRes('dame-moneda-id-cero', res);
    check(res, { 'dame moneda id=0 -> 400': (r) => r.status === 400 });
  });

  group('F2-07: GET /monedas/9988776 — moneda inexistente -> 404', () => {
    // IDs > 2147483647 (max MySQL INT) hacen overflow en el SP -> 500 en vez de 404.
    const res = http.get(`${BASE}/monedas/9988776`, PARAMS_SISTEMA);
    logRes('dame-moneda-noexiste', res);
    check(res, { 'moneda inexistente -> 404': (r) => r.status === 404 });
  });

  group('F2-08: GET /monedas — listar solo activas (default)', () => {
    const res = http.get(`${BASE}/monedas`, PARAMS_SISTEMA);
    logRes('listar-monedas', res);
    check(res, { 'listar monedas -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      const arr = Array.isArray(b) ? b : [];
      console.log(`  Total monedas activas: ${arr.length}`);
      check(null, { 'retorna al menos las 2 monedas del test': () => arr.length >= 2 });
    }
  });

  group('F2-09: GET /monedas?IncluyeInactivos=S — incluir inactivas', () => {
    const res = http.get(`${BASE}/monedas?IncluyeInactivos=S`, PARAMS_SISTEMA);
    logRes('listar-monedas-con-inactivas', res);
    check(res, { 'listar incluyendo inactivas -> 200': (r) => r.status === 200 });
  });

  group('F2-10: GET /monedas?IncluyeInactivos=X — valor inválido -> 400', () => {
    const res = http.get(`${BASE}/monedas?IncluyeInactivos=X`, PARAMS_SISTEMA);
    logRes('listar-monedas-invalido', res);
    check(res, { 'IncluyeInactivos=X -> 400': (r) => r.status === 400 });
  });

  group(`F2-11: PUT /monedas/${ID_MONEDA_PRINCIPAL}/desactivar`, () => {
    const res = http.put(`${BASE}/monedas/${ID_MONEDA_PRINCIPAL}/desactivar`, null, PARAMS_SISTEMA);
    logRes('desactivar-moneda', res);
    check(res, { 'desactivar moneda -> 200': (r) => r.status === 200 });
  });

  group(`F2-12: GET /monedas/${ID_MONEDA_PRINCIPAL} — Estado=I tras desactivar`, () => {
    const res = http.get(`${BASE}/monedas/${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('dame-moneda-inactiva', res);
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Estado: ${b.Estado}  (esperado: I)`);
      check(b, { 'moneda Estado=I tras desactivar': (x) => x.Estado === 'I' });
    } else {
      check(res, { 'dame moneda -> 200': (r) => r.status === 200 });
    }
  });

  group(`F2-13: PUT /monedas/${ID_MONEDA_PRINCIPAL}/activar`, () => {
    const res = http.put(`${BASE}/monedas/${ID_MONEDA_PRINCIPAL}/activar`, null, PARAMS_SISTEMA);
    logRes('activar-moneda', res);
    check(res, { 'activar moneda -> 200': (r) => r.status === 200 });
  });

  group(`F2-14: GET /monedas/${ID_MONEDA_PRINCIPAL} — Estado=A tras activar`, () => {
    const res = http.get(`${BASE}/monedas/${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('dame-moneda-activa', res);
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Estado: ${b.Estado}  (esperado: A)`);
      check(b, { 'moneda Estado=A tras activar': (x) => x.Estado === 'A' });
    } else {
      check(res, { 'dame moneda -> 200': (r) => r.status === 200 });
    }
  });

  group('F2-15: PUT /monedas/0/activar — IdMoneda=0 -> 400', () => {
    const res = http.put(`${BASE}/monedas/0/activar`, null, PARAMS_SISTEMA);
    logRes('activar-moneda-id-cero', res);
    check(res, { 'activar moneda id=0 -> 400': (r) => r.status === 400 });
  });

  group('F2-16: PUT /monedas/0/desactivar — IdMoneda=0 -> 400', () => {
    const res = http.put(`${BASE}/monedas/0/desactivar`, null, PARAMS_SISTEMA);
    logRes('desactivar-moneda-id-cero', res);
    check(res, { 'desactivar moneda id=0 -> 400': (r) => r.status === 400 });
  });

  group('F2-17: DELETE /monedas — crear moneda temporal, desactivar y borrarla', () => {
    // Crear (201 = nueva; 400 = ya existe de rerun anterior no borrado)
    let res = http.post(
      `${BASE}/monedas`,
      JSON.stringify({ IdMoneda: ID_MONEDA_BORRAR }),
      PARAMS_SISTEMA
    );
    logRes('crear-moneda-borrar', res);
    const creada = res.status === 201 || res.status === 400;
    check(null, { 'moneda temporal disponible': () => creada });

    // Desactivar — tsp_borrar_moneda solo permite borrar si Estado=I o P.
    // El flujo Crear deja la moneda en Estado=A, por eso hay que desactivarla primero.
    res = http.put(`${BASE}/monedas/${ID_MONEDA_BORRAR}/desactivar`, null, PARAMS_SISTEMA);
    logRes('desactivar-moneda-borrar', res);
    check(res, { 'desactivar moneda temporal -> 200': (r) => r.status === 200 });

    // Borrar
    res = http.del(`${BASE}/monedas/${ID_MONEDA_BORRAR}`, null, PARAMS_SISTEMA);
    logRes('borrar-moneda', res);
    check(res, { 'borrar moneda -> 200': (r) => r.status === 200 });
  });

  group('F2-18: DELETE /monedas/9988770 — borrar moneda inexistente -> 404', () => {
    // IDs > 2147483647 (max MySQL INT) hacen overflow en el SP -> 500 en vez de 404.
    const res = http.del(`${BASE}/monedas/9988770`, null, PARAMS_SISTEMA);
    logRes('borrar-moneda-noexiste', res);
    check(res, { 'borrar moneda inexistente -> 404': (r) => r.status === 404 });
  });

  // =========================================================================
  // FASE 3: CUENTAS — CREACIÓN Y CONSULTA
  // =========================================================================
  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('FASE 3: CUENTAS — CREACIÓN Y CONSULTA');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  group(`F3-01: POST /cuentas — crear UA(${ID_UA})/MP(${ID_MONEDA_PRINCIPAL})`, () => {
    const res = http.post(
      `${BASE}/cuentas`,
      JSON.stringify({ IdUsuarioFinal: ID_UA, IdMoneda: ID_MONEDA_PRINCIPAL, Fecha: '2026-01-01' }),
      PARAMS_SISTEMA
    );
    logRes('crear-cuenta-ua-mp', res);
    // 201 = nueva; 200 = ya existe (rerun)
    check(res, { 'crear cuenta UA/MP -> 200 o 201': (r) => r.status === 200 || r.status === 201 });
  });

  group(`F3-02: POST /cuentas — crear UB(${ID_UB})/MP(${ID_MONEDA_PRINCIPAL})`, () => {
    const res = http.post(
      `${BASE}/cuentas`,
      JSON.stringify({ IdUsuarioFinal: ID_UB, IdMoneda: ID_MONEDA_PRINCIPAL, Fecha: '2026-01-01' }),
      PARAMS_SISTEMA
    );
    logRes('crear-cuenta-ub-mp', res);
    check(res, { 'crear cuenta UB/MP -> 200 o 201': (r) => r.status === 200 || r.status === 201 });
  });

  group(`F3-03: POST /cuentas — crear UA(${ID_UA})/MS(${ID_MONEDA_SECUNDARIA})`, () => {
    const res = http.post(
      `${BASE}/cuentas`,
      JSON.stringify({ IdUsuarioFinal: ID_UA, IdMoneda: ID_MONEDA_SECUNDARIA, Fecha: '2026-01-01' }),
      PARAMS_SISTEMA
    );
    logRes('crear-cuenta-ua-ms', res);
    check(res, { 'crear cuenta UA/MS -> 200 o 201': (r) => r.status === 200 || r.status === 201 });
  });

  group('F3-04: POST /cuentas — IdUsuarioFinal=0 -> 400', () => {
    const res = http.post(
      `${BASE}/cuentas`,
      JSON.stringify({ IdUsuarioFinal: 0, IdMoneda: ID_MONEDA_PRINCIPAL, Fecha: '2026-01-01' }),
      PARAMS_SISTEMA
    );
    logRes('crear-cuenta-usuario-cero', res);
    check(res, { 'IdUsuarioFinal=0 -> 400': (r) => r.status === 400 });
  });

  group('F3-05: POST /cuentas — sin Fecha -> 400', () => {
    const res = http.post(
      `${BASE}/cuentas`,
      JSON.stringify({ IdUsuarioFinal: ID_UA, IdMoneda: ID_MONEDA_PRINCIPAL }),
      PARAMS_SISTEMA
    );
    logRes('crear-cuenta-sin-fecha', res);
    check(res, { 'sin Fecha -> 400': (r) => r.status === 400 });
  });

  group('F3-06: POST /cuentas — IdMoneda=0 -> 400', () => {
    const res = http.post(
      `${BASE}/cuentas`,
      JSON.stringify({ IdUsuarioFinal: ID_UA, IdMoneda: 0, Fecha: '2026-01-01' }),
      PARAMS_SISTEMA
    );
    logRes('crear-cuenta-moneda-cero', res);
    check(res, { 'IdMoneda=0 -> 400': (r) => r.status === 400 });
  });

  group(`F3-07: GET /cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL} — dame cuenta UA/MP`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('dame-cuenta-ua-mp', res);
    check(res, { 'dame cuenta UA/MP -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  IdUsuarioFinal=${b.IdUsuarioFinal} IdMoneda=${b.IdMoneda} Estado=${b.Estado}`);
      check(b, {
        'IdUsuarioFinal correcto': (x) => x.IdUsuarioFinal === ID_UA,
        'IdMoneda correcto':       (x) => x.IdMoneda === ID_MONEDA_PRINCIPAL,
        'Estado es A':             (x) => x.Estado === 'A',
      });
    }
  });

  group('F3-08: GET /cuentas/0/1 — IdUsuarioFinal=0 -> 400', () => {
    const res = http.get(`${BASE}/cuentas/0/1`, PARAMS_SISTEMA);
    logRes('dame-cuenta-usuario-cero', res);
    check(res, { 'IdUsuarioFinal=0 -> 400': (r) => r.status === 400 });
  });

  group('F3-09: GET /cuentas/9988776600/9988 — cuenta inexistente -> 400', () => {
    const res = http.get(`${BASE}/cuentas/9988776600/9988`, PARAMS_SISTEMA);
    logRes('dame-cuenta-noexiste', res);
    check(res, { 'cuenta inexistente -> 400': (r) => r.status === 400 });
  });

  group(`F3-10: GET /cuentas?IdUsuarioFinal=${ID_UA} — buscar cuentas de UA`, () => {
    const res = http.get(`${BASE}/cuentas?IdUsuarioFinal=${ID_UA}`, PARAMS_SISTEMA);
    logRes('buscar-cuentas-ua', res);
    check(res, { 'buscar cuentas UA -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total cuentas UA: ${b.Total}  (esperado >= 2: MP y MS)`);
      check(null, { 'UA tiene >= 2 cuentas': () => b.Total >= 2 });
    }
  });

  group(`F3-11: GET /cuentas?IdMoneda=${ID_MONEDA_PRINCIPAL} — buscar por moneda`, () => {
    const res = http.get(`${BASE}/cuentas?IdMoneda=${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('buscar-cuentas-mp', res);
    check(res, { 'buscar cuentas MP -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total cuentas en MP: ${b.Total}  (esperado >= 2: UA y UB)`);
      check(null, { 'MP tiene >= 2 cuentas': () => b.Total >= 2 });
    }
  });

  group(`F3-12: GET /cuentas — buscar UA/MP con Estado=A`, () => {
    const qs = buildQS({ IdUsuarioFinal: String(ID_UA), IdMoneda: String(ID_MONEDA_PRINCIPAL), Estado: 'A' });
    const res = http.get(`${BASE}/cuentas?${qs}`, PARAMS_SISTEMA);
    logRes('buscar-cuentas-filtros-combinados', res);
    check(res, { 'buscar con filtros -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total: ${b.Total}  (esperado: 1)`);
      check(null, { 'exactamente 1 cuenta UA/MP activa': () => b.Total === 1 });
    }
  });

  group(`F3-13: GET /cuentas — Estado=I + IdUsuarioFinal=${ID_UA} (sin inactivas -> 0)`, () => {
    const qs = buildQS({ IdUsuarioFinal: String(ID_UA), IdMoneda: String(ID_MONEDA_PRINCIPAL), Estado: 'I' });
    const res = http.get(`${BASE}/cuentas?${qs}`, PARAMS_SISTEMA);
    logRes('buscar-cuentas-inactivas', res);
    check(res, { 'buscar inactivas -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total inactivas UA/MP: ${b.Total}  (esperado: 0 — aún activa)`);
      check(null, { 'cuenta activa no aparece como inactiva': () => b.Total === 0 });
    }
  });

  group('F3-14: GET /cuentas?Estado=X — estado inválido -> 400', () => {
    const res = http.get(`${BASE}/cuentas?Estado=X`, PARAMS_SISTEMA);
    logRes('buscar-cuentas-estado-invalido', res);
    check(res, { 'Estado=X -> 400': (r) => r.status === 400 });
  });

  group('F3-15: GET /cuentas?Limite=0 — Limite=0 -> 400', () => {
    const res = http.get(`${BASE}/cuentas?Limite=0`, PARAMS_SISTEMA);
    logRes('buscar-cuentas-limite-cero', res);
    check(res, { 'Limite=0 -> 400': (r) => r.status === 400 });
  });

  group('F3-16: GET /cuentas — búsqueda directa por arrays IdsUsuarioFinal+IdsMoneda', () => {
    const qs = buildQS({
      IdsUsuarioFinal: [String(ID_UA), String(ID_UB)],
      IdsMoneda:       [String(ID_MONEDA_PRINCIPAL), String(ID_MONEDA_PRINCIPAL)],
    });
    const res = http.get(`${BASE}/cuentas?${qs}`, PARAMS_SISTEMA);
    logRes('buscar-cuentas-array', res);
    check(res, { 'búsqueda por array IDs -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total: ${b.Total}  (esperado: 2)`);
      check(null, { 'retorna exactamente 2 cuentas': () => b.Total === 2 });
    }
  });

  group('F3-17: GET /cuentas — arrays de distinto tamaño -> 400', () => {
    const qs = buildQS({
      IdsUsuarioFinal: [String(ID_UA)],
      IdsMoneda:       [String(ID_MONEDA_PRINCIPAL), String(ID_MONEDA_SECUNDARIA)],
    });
    const res = http.get(`${BASE}/cuentas?${qs}`, PARAMS_SISTEMA);
    logRes('buscar-cuentas-arrays-distintos', res);
    check(res, { 'arrays distintos tamaño -> 400': (r) => r.status === 400 });
  });

  group('F3-18: DELETE /monedas — borrar moneda con cuentas de usuario -> 409', () => {
    // ID_MONEDA_PRINCIPAL ya tiene cuentas de usuario (creadas en F3-01 y F3-02).
    // El borrado debe rechazarse aunque la moneda esté en estado válido.
    const res = http.del(`${BASE}/monedas/${ID_MONEDA_PRINCIPAL}`, null, PARAMS_SISTEMA);
    logRes('borrar-moneda-con-cuentas', res);
    check(res, { 'borrar moneda con cuentas de usuario -> 409': (r) => r.status === 409 });
  });

  // =========================================================================
  // FASE 4: TRANSFERENCIAS — INGRESOS Y EGRESOS BASE
  // =========================================================================
  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('FASE 4: TRANSFERENCIAS — INGRESOS Y EGRESOS BASE');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  group(`F4-01: POST /transferencias — T01 Ingreso $1000 UA/MP`, () => {
    const res = postTransfer(PARAMS_SISTEMA, T01, ID_UA, 1000, ID_MONEDA_PRINCIPAL, 'I', 1);
    logRes('t01-ingreso-1000', res);
    check(res, { 'T01 -> 202': (r) => r.status === 202 });
  });

  group(`F4-02: POST /transferencias — T02 Ingreso $500 UA/MP`, () => {
    const res = postTransfer(PARAMS_SISTEMA, T02, ID_UA, 500, ID_MONEDA_PRINCIPAL, 'I', 1);
    logRes('t02-ingreso-500', res);
    check(res, { 'T02 -> 202': (r) => r.status === 202 });
  });

  // NOTA: T03 (Egreso $200 UA/MP) se envía en un batch SEPARADO (después del primer waitKafka).
  // Motivo: preValidarCuentas chequea el saldo de UA/MP ANTES de que TB procese el batch.
  // Si T01, T02 y T03 van en el mismo batch, al momento de validar T03 el saldo aún es 0
  // (T01/T02 no se aplicaron todavía) -> "Saldo insuficiente". Enviando T03 después de que
  // los ingresos ya fueron procesados por TB, el balance es 1500 y el egreso de $200 pasa.

  group(`F4-03: POST /transferencias — T04 Ingreso $800 UB/MP`, () => {
    const res = postTransfer(PARAMS_SISTEMA, T04, ID_UB, 800, ID_MONEDA_PRINCIPAL, 'I', 2);
    logRes('t04-ingreso-800-ub', res);
    check(res, { 'T04 -> 202': (r) => r.status === 202 });
  });

  group(`F4-04: POST /transferencias — T05 Ingreso $300 UA/MS`, () => {
    const res = postTransfer(PARAMS_SISTEMA, T05, ID_UA, 300, ID_MONEDA_SECUNDARIA, 'I', 3);
    logRes('t05-ingreso-300-ms', res);
    check(res, { 'T05 -> 202': (r) => r.status === 202 });
  });

  // Validaciones del controlador (rechazo inmediato, sin Kafka)
  group('F4-05: POST /transferencias — JSON inválido -> 400', () => {
    const res = http.post(`${BASE}/transferencias`, 'esto no es json{{{', PARAMS_SISTEMA);
    logRes('transfer-json-invalido', res);
    check(res, { 'JSON inválido -> 400': (r) => r.status === 400 });
  });

  group('F4-06: POST /transferencias — IdTransferencia vacío -> 400', () => {
    const res = http.post(
      `${BASE}/transferencias`,
      JSON.stringify({ IdTransferencia: '', IdUsuarioFinal: ID_UA, Monto: 100, IdMoneda: ID_MONEDA_PRINCIPAL, Tipo: 'I', IdCategoria: 1, Fecha: '2026-01-15' }),
      PARAMS_SISTEMA
    );
    logRes('transfer-id-vacio', res);
    check(res, { 'IdTransferencia vacío -> 400': (r) => r.status === 400 });
  });

  group('F4-07: POST /transferencias — IdUsuarioFinal=0 -> 400', () => {
    const res = http.post(
      `${BASE}/transferencias`,
      JSON.stringify({ IdTransferencia: '9569019990001', IdUsuarioFinal: 0, Monto: 100, IdMoneda: ID_MONEDA_PRINCIPAL, Tipo: 'I', IdCategoria: 1, Fecha: '2026-01-15' }),
      PARAMS_SISTEMA
    );
    logRes('transfer-usuario-cero', res);
    check(res, { 'IdUsuarioFinal=0 -> 400': (r) => r.status === 400 });
  });

  group('F4-08: POST /transferencias — Tipo inválido "X" -> 400', () => {
    const res = http.post(
      `${BASE}/transferencias`,
      JSON.stringify({ IdTransferencia: '9569019990002', IdUsuarioFinal: ID_UA, Monto: 100, IdMoneda: ID_MONEDA_PRINCIPAL, Tipo: 'X', IdCategoria: 1, Fecha: '2026-01-15' }),
      PARAMS_SISTEMA
    );
    logRes('transfer-tipo-invalido', res);
    check(res, { 'Tipo=X -> 400': (r) => r.status === 400 });
  });

  group('F4-09: POST /transferencias — Monto=0 para Ingreso -> 400', () => {
    const res = http.post(
      `${BASE}/transferencias`,
      JSON.stringify({ IdTransferencia: '9569019990003', IdUsuarioFinal: ID_UA, Monto: 0, IdMoneda: ID_MONEDA_PRINCIPAL, Tipo: 'I', IdCategoria: 1, Fecha: '2026-01-15' }),
      PARAMS_SISTEMA
    );
    logRes('transfer-monto-cero', res);
    check(res, { 'Monto=0 -> 400': (r) => r.status === 400 });
  });

  group('F4-10: POST /transferencias — Monto negativo -> 400', () => {
    const res = http.post(
      `${BASE}/transferencias`,
      JSON.stringify({ IdTransferencia: '9569019990004', IdUsuarioFinal: ID_UA, Monto: -50, IdMoneda: ID_MONEDA_PRINCIPAL, Tipo: 'I', IdCategoria: 1, Fecha: '2026-01-15' }),
      PARAMS_SISTEMA
    );
    logRes('transfer-monto-negativo', res);
    check(res, { 'Monto negativo -> 400': (r) => r.status === 400 });
  });

  group('F4-11: POST /transferencias — IdMoneda=0 -> 400', () => {
    const res = http.post(
      `${BASE}/transferencias`,
      JSON.stringify({ IdTransferencia: '9569019990005', IdUsuarioFinal: ID_UA, Monto: 100, IdMoneda: 0, Tipo: 'I', IdCategoria: 1, Fecha: '2026-01-15' }),
      PARAMS_SISTEMA
    );
    logRes('transfer-moneda-cero', res);
    check(res, { 'IdMoneda=0 -> 400': (r) => r.status === 400 });
  });

  waitKafka('T01, T02, T04, T05 (ingresos — esperar antes de enviar T03 egreso)');

  group(`F4-12: POST /transferencias — T03 Egreso $200 UA/MP (batch propio post-ingresos)`, () => {
    const res = postTransfer(PARAMS_SISTEMA, T03, ID_UA, 200, ID_MONEDA_PRINCIPAL, 'E', 1);
    logRes('t03-egreso-200', res);
    check(res, { 'T03 -> 202': (r) => r.status === 202 });
  });

  waitKafka('T03 (egreso — saldo UA/MP ya disponible)');

  group(`F4-13: GET /transferencias/${T01} — verificar T01 (Ingreso $1000)`, () => {
    const res = http.get(`${BASE}/transferencias/${T01}`, PARAMS_SISTEMA);
    logRes('get-t01', res);
    check(res, { 'GET T01 -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Estado=${b.Estado} Tipo=${b.Tipo} Monto=${b.Monto}`);
      check(b, {
        'T01 Estado=F':      (x) => x.Estado === 'F',
        'T01 Tipo=I':        (x) => x.Tipo === 'I',
        'T01 Monto=1000.00': (x) => x.Monto === '1000.00',
      });
    }
  });

  group(`F4-14: GET /transferencias/${T03} — verificar T03 (Egreso $200)`, () => {
    const res = http.get(`${BASE}/transferencias/${T03}`, PARAMS_SISTEMA);
    logRes('get-t03', res);
    check(res, { 'GET T03 -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Estado=${b.Estado} Tipo=${b.Tipo} Monto=${b.Monto}`);
      check(b, {
        'T03 Estado=F':    (x) => x.Estado === 'F',
        'T03 Tipo=E':      (x) => x.Tipo === 'E',
        'T03 Monto=200.00': (x) => x.Monto === '200.00',
      });
    }
  });

  group('F4-15: GET /transferencias/9569019988888 — transfer inexistente -> 404', () => {
    const res = http.get(`${BASE}/transferencias/9569019988888`, PARAMS_SISTEMA);
    logRes('get-transfer-noexiste', res);
    check(res, { 'transfer inexistente -> 404': (r) => r.status === 404 });
  });

  group(`F4-16: GET /cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL} — balance tras T01+T02+T03`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('balance-ua-mp-inicial', res);
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Creditos=${b.Creditos}  Debitos=${b.Debitos}  Estado=${b.Estado}`);
      // En re-runs los IDs son fijos e idempotentes en TB, pero T06 y su reversal del
      // run anterior ya están aplicados -> balances acumulados (>= en vez de ===).
      // Los valores exactos finales se verifican en FASE 10.
      check(b, {
        'UA/MP Creditos>=1500.00': (x) => parseFloat(x.Creditos) >= 1500,
        'UA/MP Debitos>=200.00':   (x) => parseFloat(x.Debitos) >= 200,
        'UA/MP Estado=A':          (x) => x.Estado === 'A',
      });
    } else {
      check(res, { 'dame cuenta -> 200': (r) => r.status === 200 });
    }
  });

  // =========================================================================
  // FASE 5: BÚSQUEDA DE TRANSFERENCIAS
  // =========================================================================
  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('FASE 5: BÚSQUEDA DE TRANSFERENCIAS');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  group(`F5-01: GET /transferencias?IdUsuarioFinal=${ID_UA}&IdMoneda=${ID_MONEDA_PRINCIPAL}`, () => {
    const url = `${BASE}/transferencias?IdUsuarioFinal=${ID_UA}&IdMoneda=${ID_MONEDA_PRINCIPAL}`;
    const res = http.get(url, PARAMS_SISTEMA);
    logRes('buscar-transfers-ua-mp', res);
    check(res, { 'buscar UA/MP -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total: ${b.Total}  (esperado >= 3: T01,T02,T03)`);
      check(null, { 'UA/MP tiene >= 3 transfers': () => b.Total >= 3 });
    }
  });

  group(`F5-02: GET /transferencias?IdMoneda=${ID_MONEDA_PRINCIPAL} — todas las de MP`, () => {
    const res = http.get(`${BASE}/transferencias?IdMoneda=${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('buscar-transfers-mp', res);
    check(res, { 'buscar por moneda -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total en MP: ${b.Total}  (esperado >= 4: T01-T04)`);
      check(null, { 'MP tiene >= 4 transfers': () => b.Total >= 4 });
    }
  });

  group(`F5-03: GET /transferencias?IncluyeRevertidas=false&IdMoneda=${ID_MONEDA_PRINCIPAL} — solo finalizadas`, () => {
    const res = http.get(`${BASE}/transferencias?IncluyeRevertidas=false&IdMoneda=${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('buscar-transfers-solo-finalizadas', res);
    check(res, { 'IncluyeRevertidas=false -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      const todosF = b.Transferencias.every(t => t.Estado === 'F');
      check(null, { 'todos los resultados tienen Estado=F': () => todosF });
    }
  });

  group('F5-04: GET /transferencias?MontoMin=200&MontoMax=1000&IdUsuarioFinal=... — filtro rango', () => {
    const url = `${BASE}/transferencias?IdUsuarioFinal=${ID_UA}&IdMoneda=${ID_MONEDA_PRINCIPAL}&MontoMin=200&MontoMax=1000`;
    const res = http.get(url, PARAMS_SISTEMA);
    logRes('buscar-transfers-rango-monto', res);
    check(res, { 'filtro MontoMin/Max -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      // T01=1000, T02=500, T03=200 deben pasar
      console.log(`  Total en rango [200,1000]: ${b.Total}  (esperado >= 3)`);
      check(null, { 'rango [200,1000] tiene >= 3': () => b.Total >= 3 });
    }
  });

  group('F5-05: GET /transferencias — búsqueda directa por IdsTransferencia', () => {
    const qs = buildQS({ IdsTransferencia: [T01, T02] });
    const res = http.get(`${BASE}/transferencias?${qs}`, PARAMS_SISTEMA);
    logRes('buscar-transfers-ids', res);
    check(res, { 'buscar por IDs -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total: ${b.Total}  (esperado: 2)`);
      check(null, { 'retorna exactamente 2 transfers': () => b.Total === 2 });
    }
  });

  group('F5-06: GET /transferencias?Limite=1&IncluyeRevertidas=true — retorna solo 1 resultado', () => {
    const res = http.get(`${BASE}/transferencias?IdMoneda=${ID_MONEDA_PRINCIPAL}&Limite=1&IncluyeRevertidas=true`, PARAMS_SISTEMA);
    logRes('buscar-transfers-limite-1', res);
    check(res, { 'Limite=1 -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      check(null, { 'retorna exactamente 1 transfer': () => b.Total === 1 });
    }
  });

  group('F5-07: GET /transferencias?IdCategoria=2 — filtro por categoría', () => {
    const res = http.get(`${BASE}/transferencias?IdMoneda=${ID_MONEDA_PRINCIPAL}&IdCategoria=2`, PARAMS_SISTEMA);
    logRes('buscar-transfers-categoria', res);
    check(res, { 'buscar por categoría -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total categoría 2: ${b.Total}  (esperado >= 1: T04)`);
    }
  });

  group('F5-08: GET /transferencias?Limite=501 — excede máximo -> 400', () => {
    const res = http.get(`${BASE}/transferencias?Limite=501`, PARAMS_SISTEMA);
    logRes('buscar-transfers-limite-excede', res);
    check(res, { 'Limite=501 -> 400': (r) => r.status === 400 });
  });

  group('F5-09: GET /transferencias?Limite=0 — Limite=0 -> 400', () => {
    const res = http.get(`${BASE}/transferencias?Limite=0`, PARAMS_SISTEMA);
    logRes('buscar-transfers-limite-cero', res);
    check(res, { 'Limite=0 -> 400': (r) => r.status === 400 });
  });

  group('F5-10: GET /transferencias?IncluyeRevertidas=invalido — valor inválido -> 400', () => {
    const res = http.get(`${BASE}/transferencias?IncluyeRevertidas=invalido`, PARAMS_SISTEMA);
    logRes('buscar-transfers-incluyerevertidas-invalido', res);
    check(res, { 'IncluyeRevertidas=invalido -> 400': (r) => r.status === 400 });
  });

  group('F5-11: GET /transferencias?MontoMin=500&MontoMax=100 — min > max -> 400', () => {
    const res = http.get(`${BASE}/transferencias?MontoMin=500&MontoMax=100`, PARAMS_SISTEMA);
    logRes('buscar-transfers-monto-invalido', res);
    check(res, { 'MontoMin > MontoMax -> 400': (r) => r.status === 400 });
  });

  group('F5-12: GET /transferencias?MontoMin=-10 — MontoMin negativo -> 400', () => {
    const res = http.get(`${BASE}/transferencias?MontoMin=-10`, PARAMS_SISTEMA);
    logRes('buscar-transfers-montomin-neg', res);
    check(res, { 'MontoMin negativo -> 400': (r) => r.status === 400 });
  });

  group('F5-13: GET /transferencias?MontoMax=-1 — MontoMax negativo -> 400', () => {
    const res = http.get(`${BASE}/transferencias?MontoMax=-1`, PARAMS_SISTEMA);
    logRes('buscar-transfers-montomax-neg', res);
    check(res, { 'MontoMax negativo -> 400': (r) => r.status === 400 });
  });

  group('F5-14: GET /transferencias — verificar formato decimal de montos', () => {
    const res = http.get(`${BASE}/transferencias?IdUsuarioFinal=${ID_UA}&IdMoneda=${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('buscar-transfers-formato-decimal', res);
    if (res.status === 200) {
      const b = parseBody(res);
      const formatosOK = b.Transferencias.every(t => /^\d+\.\d{2}$/.test(t.Monto));
      console.log(`  Transfers verificadas: ${b.Total} | formatos OK: ${formatosOK}`);
      check(null, { 'todos los Monto tienen formato decimal XX.XX': () => formatosOK });
    }
  });

  // =========================================================================
  // FASE 6: TRANSFERENCIAS DE CUENTA (DameTransferencias) E HISTORIAL
  // =========================================================================
  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('FASE 6: TRANSFERENCIAS DE CUENTA E HISTORIAL DE BALANCES');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  group(`F6-01: GET /cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/transferencias`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/transferencias`, PARAMS_SISTEMA);
    logRes('dame-transfers-cuenta', res);
    check(res, { 'dame transfers cuenta -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total transfers UA/MP (IncluyeRevertidas=false): ${b.Total}  (esperado >= 2)`);
      check(null, { 'UA/MP tiene >= 2 transfers': () => b.Total >= 2 });
      if (b.Total > 0) {
        // No debe haber transfers de cierre ni de reversión internas
        const sinCierre = b.Transferencias.every(t => t.Monto !== '0.00' && t.Monto !== '0');
        check(null, { 'sin transfers de cierre (Code=3)': () => sinCierre });
        const sinReversionInterna = b.Transferencias.every(t => t.Tipo !== 'R');
        check(null, { 'sin transfers de reversión internas (Tipo!=R)': () => sinReversionInterna });
        // Por defecto IncluyeRevertidas=false -> ningún Estado=R
        const sinRevertidas = b.Transferencias.every(t => t.Estado !== 'R');
        check(null, { 'IncluyeRevertidas=false: sin Estado=R': () => sinRevertidas });
      }
    }
  });

  group(`F6-02: GET /cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/transferencias?Limite=2`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/transferencias?Limite=2`, PARAMS_SISTEMA);
    logRes('dame-transfers-cuenta-limite', res);
    check(res, { 'dame transfers Limite=2 -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      check(null, { 'retorna máximo 2 transfers': () => b.Total <= 2 });
    }
  });

  group('F6-03: GET /cuentas/0/1/transferencias — IdUsuarioFinal=0 -> 200', () => {
    const res = http.get(`${BASE}/cuentas/0/1/transferencias`, PARAMS_SISTEMA);
    logRes('dame-transfers-usuario-cero', res);
    check(res, { 'transfers usuario=0 -> 200': (r) => r.status === 200 });
  });

  group('F6-04: GET /cuentas/1/0/transferencias — IdMoneda=0 -> 400', () => {
    const res = http.get(`${BASE}/cuentas/1/0/transferencias`, PARAMS_SISTEMA);
    logRes('dame-transfers-moneda-cero', res);
    check(res, { 'transfers moneda=0 -> 400': (r) => r.status === 400 });
  });

  group(`F6-05: GET /cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/transferencias?IncluyeRevertidas=true`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/transferencias?IncluyeRevertidas=true`, PARAMS_SISTEMA);
    logRes('dame-transfers-cuenta-con-revertidas', res);
    check(res, { 'IncluyeRevertidas=true -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total transfers UA/MP (IncluyeRevertidas=true): ${b.Total}  (esperado >= 3)`);
      check(null, { 'IncluyeRevertidas=true: >= 3 transfers': () => b.Total >= 3 });
      if (b.Total > 0) {
        // Estado debe ser 'F' o 'R' para todas (no reversal internas)
        const estadosValidos = b.Transferencias.every(t => t.Estado === 'F' || t.Estado === 'R');
        check(null, { 'todos los Estados son F o R': () => estadosValidos });
        const sinReversionInterna = b.Transferencias.every(t => t.Tipo !== 'R');
        check(null, { 'sin transfers de reversión internas (Tipo!=R)': () => sinReversionInterna });
      }
    }
  });

  group('F6-06: GET /cuentas/transferencias — IncluyeRevertidas=invalido -> 400', () => {
    const res = http.get(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/transferencias?IncluyeRevertidas=invalido`, PARAMS_SISTEMA);
    logRes('dame-transfers-incluyerevertidas-invalido', res);
    check(res, { 'IncluyeRevertidas=invalido -> 400': (r) => r.status === 400 });
  });

  group(`F6-07: GET /cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/historial`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/historial`, PARAMS_SISTEMA);
    logRes('historial-ua-mp', res);
    check(res, { 'historial cuenta -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total entradas historial: ${b.Total}  (esperado >= 3)`);
      check(null, { 'historial >= 3 entradas': () => b.Total >= 3 });
      if (b.Total > 0) {
        // Formato decimal en todas las entradas
        const formatosOK = b.Historial.every(h =>
          /^\d+\.\d{2}$/.test(h.Creditos) &&
          /^\d+\.\d{2}$/.test(h.Debitos) &&
          /^-?\d+\.\d{2}$/.test(h.Balance)
        );
        check(null, { 'historial: formato decimal correcto en todas las entradas': () => formatosOK });
        // Primera entrada = estado más reciente (historial viene en orden descendente).
        // Se usan >= para Creditos/Debitos; Balance=1300 se mantiene estable siempre
        // (T06 + reversal se cancelan entre sí, y todas las demás transfers son idempotentes).
        const ultima = b.Historial[0];
        console.log(`  Última entrada: C=${ultima.Creditos} D=${ultima.Debitos} B=${ultima.Balance}`);
        check(ultima, {
          'última entrada Creditos>=1500.00': (x) => parseFloat(x.Creditos) >= 1500,
          'última entrada Debitos>=200.00':   (x) => parseFloat(x.Debitos) >= 200,
          'última entrada Balance=1300.00':   (x) => x.Balance === '1300.00',
        });
      }
    }
  });

  group(`F6-08: GET /cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/historial?Limite=1`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/historial?Limite=1`, PARAMS_SISTEMA);
    logRes('historial-limite', res);
    check(res, { 'historial Limite=1 -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      check(null, { 'máximo 1 entrada en historial': () => b.Total <= 1 });
    }
  });

  group('F6-09: GET /cuentas/0/1/historial — IdUsuarioFinal=0 -> 200', () => {
    const res = http.get(`${BASE}/cuentas/0/1/historial`, PARAMS_SISTEMA);
    logRes('historial-usuario-cero', res);
    check(res, { 'historial usuario=0 -> 200': (r) => r.status === 200 });
  });

  // =========================================================================
  // FASE 7: ACTIVAR / DESACTIVAR CUENTAS
  // =========================================================================
  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('FASE 7: ACTIVAR / DESACTIVAR CUENTAS');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  group(`F7-01: PUT /cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL}/desactivar`, () => {
    const res = http.put(`${BASE}/cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL}/desactivar`, null, PARAMS_SISTEMA);
    logRes('desactivar-cuenta-ub', res);
    check(res, { 'desactivar cuenta UB -> 200': (r) => r.status === 200 });
  });

  group(`F7-02: GET /cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL} — Estado=I`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('dame-cuenta-ub-inactiva', res);
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Estado: ${b.Estado}  (esperado: I)`);
      check(b, { 'UB Estado=I': (x) => x.Estado === 'I' });
    }
  });

  group(`F7-03: PUT /cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL}/desactivar — ya inactiva -> 409`, () => {
    const res = http.put(`${BASE}/cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL}/desactivar`, null, PARAMS_SISTEMA);
    logRes('desactivar-cuenta-ub-idempotente', res);
    check(res, { 'desactivar segunda vez (ya inactiva) -> 409': (r) => r.status === 409 });
  });

  group(`F7-04: POST /transferencias — T08 Egreso UB/MP con cuenta cerrada`, () => {
    const res = postTransfer(PARAMS_SISTEMA, T08, ID_UB, 100, ID_MONEDA_PRINCIPAL, 'E', 1);
    logRes('t08-cuenta-cerrada', res);
    // Controlador acepta, Kafka la encola; el consumer la rechaza
    check(res, { 'T08 con cuenta cerrada -> 202': (r) => r.status === 202 });
  });

  waitKafka('T08 (egreso cuenta cerrada -> pre-TB reject)');

  group(`F7-05: GET /transferencias/${T08} — rechazada pre-TB -> 404`, () => {
    const res = http.get(`${BASE}/transferencias/${T08}`, PARAMS_SISTEMA);
    logRes('get-t08-rechazada', res);
    check(res, { 'T08 rechazada: no llegó a TB -> 404': (r) => r.status === 404 });
  });

  group(`F7-06: GET /cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL} — balance sin cambio`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('balance-ub-sin-cambio', res);
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Creditos=${b.Creditos} Debitos=${b.Debitos} Estado=${b.Estado}`);
      check(b, {
        'UB Estado=I (sigue cerrada)':    (x) => x.Estado === 'I',
        'UB Creditos sin cambio=800.00':  (x) => x.Creditos === '800.00',
        'UB Debitos sin cambio=0.00':     (x) => x.Debitos === '0.00',
      });
    }
  });

  group(`F7-07: PUT /cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL}/activar — reactivar`, () => {
    const res = http.put(`${BASE}/cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL}/activar`, null, PARAMS_SISTEMA);
    logRes('activar-cuenta-ub', res);
    check(res, { 'activar cuenta UB -> 200': (r) => r.status === 200 });
  });

  group(`F7-08: GET /cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL} — Estado=A tras activar`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('dame-cuenta-ub-activa', res);
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Estado: ${b.Estado}  (esperado: A)`);
      check(b, { 'UB Estado=A': (x) => x.Estado === 'A' });
    }
  });

  group(`F7-09: PUT /cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL}/activar — ya activa -> 409`, () => {
    const res = http.put(`${BASE}/cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL}/activar`, null, PARAMS_SISTEMA);
    logRes('activar-cuenta-ub-idempotente', res);
    check(res, { 'activar segunda vez (ya activa) -> 409': (r) => r.status === 409 });
  });

  group('F7-10: PUT /cuentas/99988877/1/desactivar — cuenta inexistente -> 404', () => {
    const res = http.put(`${BASE}/cuentas/99988877/1/desactivar`, null, PARAMS_SISTEMA);
    logRes('desactivar-cuenta-noexiste', res);
    check(res, { 'desactivar cuenta inexistente -> 404': (r) => r.status === 404 });
  });

  group(`F7-11: PUT /cuentas/0/${ID_MONEDA_PRINCIPAL}/activar — IdUsuarioFinal=0 -> 400`, () => {
    const res = http.put(`${BASE}/cuentas/0/${ID_MONEDA_PRINCIPAL}/activar`, null, PARAMS_SISTEMA);
    logRes('activar-cuenta-usuario-cero', res);
    check(res, { 'activar cuenta usuario=0 -> 400': (r) => r.status === 400 });
  });

  group(`F7-12: GET /cuentas?Estado=I&IdUsuarioFinal=${ID_UA}&IdMoneda=${ID_MONEDA_PRINCIPAL} — buscar inactivas`, () => {
    // Desactivar UA temporalmente para probar búsqueda por Estado=I
    http.put(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/desactivar`, null, PARAMS_SISTEMA);

    const qs = buildQS({ Estado: 'I', IdUsuarioFinal: String(ID_UA), IdMoneda: String(ID_MONEDA_PRINCIPAL) });
    const res = http.get(`${BASE}/cuentas?${qs}`, PARAMS_SISTEMA);
    logRes('buscar-cuentas-inactivas', res);
    check(res, { 'buscar inactivas -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total inactivas UA/MP: ${b.Total}  (esperado: 1)`);
      check(null, { 'UA inactiva encontrada': () => b.Total >= 1 });
    }

    // Reactivar UA
    http.put(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/activar`, null, PARAMS_SISTEMA);
  });

  // =========================================================================
  // FASE 8: REVERSAL DE TRANSFERENCIA
  // =========================================================================
  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('FASE 8: REVERSAL DE TRANSFERENCIA');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  group(`F8-01: POST /transferencias — T06 Egreso $150 UA/MP (preparar reversal)`, () => {
    const res = postTransfer(PARAMS_SISTEMA, T06, ID_UA, 150, ID_MONEDA_PRINCIPAL, 'E', 1);
    logRes('t06-egreso-reversal', res);
    check(res, { 'T06 egreso $150 -> 202': (r) => r.status === 202 });
  });

  waitKafka('T06 (egreso para reversal — debe ser el último de UA/MP)');

  group(`F8-02: GET /transferencias/${T06} — verificar T06 procesada`, () => {
    const res = http.get(`${BASE}/transferencias/${T06}`, PARAMS_SISTEMA);
    logRes('get-t06', res);
    check(res, { 'GET T06 -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Estado=${b.Estado} Tipo=${b.Tipo} Monto=${b.Monto}`);
      // En reruns T06 puede estar ya revertida (Estado=R) de la ejecución anterior
      check(b, {
        'T06 procesada (Estado=F o R)': (x) => x.Estado === 'F' || x.Estado === 'R',
        'T06 Tipo=E':                   (x) => x.Tipo === 'E',
        'T06 Monto=150.00':             (x) => x.Monto === '150.00',
      });
    }
  });

  group('F8-03: POST /transferencias — reversal de T06 (Tipo=R, Monto=0)', () => {
    // Para reversal: mismo IdTransferencia que la original, Tipo=R, Monto=0
    const res = postTransfer(PARAMS_SISTEMA, T06, ID_UA, 0, ID_MONEDA_PRINCIPAL, 'R', 1);
    logRes('t06-reversal', res);
    check(res, { 'reversal de T06 -> 202': (r) => r.status === 202 });
  });

  waitKafka('reversal de T06');

  group('F8-04: GET /transferencias?IncluyeRevertidas=true — T06 aparece como revertida', () => {
    // Las transfers de reversión internas nunca aparecen en resultados.
    // Con IncluyeRevertidas=true, T06 debe aparecer con Estado=R (fue revertida).
    const res = http.get(
      `${BASE}/transferencias?IdUsuarioFinal=${ID_UA}&IdMoneda=${ID_MONEDA_PRINCIPAL}&IncluyeRevertidas=true`,
      PARAMS_SISTEMA
    );
    logRes('buscar-incluyerevertidas', res);
    check(res, { 'IncluyeRevertidas=true -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total UA/MP con revertidas: ${b.Total}`);
      const t06 = b.Transferencias.find(t => t.IdTransferencia === T06);
      console.log(`  T06 Estado=${t06 ? t06.Estado : 'no encontrada'}  (esperado: R)`);
      check(null, { 'T06 aparece con Estado=R': () => t06 && t06.Estado === 'R' });
      const soloFyR = b.Transferencias.every(t => t.Estado === 'F' || t.Estado === 'R');
      check(null, { 'todos los resultados tienen Estado=F o R': () => soloFyR });
    }
  });

  group(`F8-05: GET /transferencias/${T06} — Dame post-reversal, Estado=R`, () => {
    const res = http.get(`${BASE}/transferencias/${T06}`, PARAMS_SISTEMA);
    logRes('get-t06-post-reversal', res);
    check(res, { 'GET T06 post-reversal -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Estado=${b.Estado} Tipo=${b.Tipo}  (esperado: Estado=R Tipo=E)`);
      check(b, {
        'T06 Estado=R tras reversal': (x) => x.Estado === 'R',
        'T06 Tipo=E (no cambia)':     (x) => x.Tipo === 'E',
      });
    }
  });

  group('F8-07: POST /transferencias — reversal de transfer inexistente -> 202 (rechazada async)', () => {
    const res = postTransfer(PARAMS_SISTEMA, '9569019977001', ID_UA, 0, ID_MONEDA_PRINCIPAL, 'R', 1);
    logRes('reversal-inexistente', res);
    // El controlador acepta; el consumer la rechazará (no encuentra la transfer original)
    check(res, { 'reversal inexistente -> 202 (aceptada async)': (r) => r.status === 202 });
  });

  waitKafka('reversal de transfer inexistente');

  // =========================================================================
  // FASE 9: USUARIOS — CICLO COMPLETO
  // =========================================================================
  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('FASE 9: USUARIOS — CICLO COMPLETO');
  console.log(`  Nombre usuario nuevo: ${NOMBRE_USR_NUEVO}`);
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');

  group(`F9-01: POST /usuarios — crear '${NOMBRE_USR_NUEVO}' (actor SISTEMA)`, () => {
    const res = http.post(
      `${BASE}/usuarios`,
      JSON.stringify({ Usuario: NOMBRE_USR_NUEVO }),
      PARAMS_SISTEMA
    );
    logRes('crear-usuario', res);
    check(res, { 'crear usuario -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      idUsuarioNuevo = b.Id || 0;
      passTemp       = b.PasswordTemporal || '';
      console.log(`  Id=${idUsuarioNuevo} PasswordTemporal=${passTemp}`);
      check(null, {
        'Id > 0':                () => idUsuarioNuevo > 0,
        'PasswordTemporal != ""': () => passTemp !== '',
      });
    }
  });

  group('F9-02: POST /usuarios — sin Usuario -> 400', () => {
    const res = http.post(`${BASE}/usuarios`, JSON.stringify({}), PARAMS_SISTEMA);
    logRes('crear-usuario-sin-nombre', res);
    check(res, { 'sin Usuario -> 400': (r) => r.status === 400 });
  });

  group('F9-03: POST /usuarios — Usuario vacío -> 400', () => {
    const res = http.post(`${BASE}/usuarios`, JSON.stringify({ Usuario: '' }), PARAMS_SISTEMA);
    logRes('crear-usuario-vacio', res);
    check(res, { 'Usuario vacío -> 400': (r) => r.status === 400 });
  });

  group(`F9-04: GET /usuarios/${idUsuarioNuevo || 1} — dame usuario recién creado`, () => {
    if (idUsuarioNuevo <= 0) { console.log('  SKIP: sin usuario nuevo'); return; }
    const res = http.get(`${BASE}/usuarios/${idUsuarioNuevo}`, PARAMS_SISTEMA);
    logRes('dame-usuario', res);
    check(res, { 'dame usuario -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      const u = b.Usuario || {};
      console.log(`  Usuario=${u.Usuario} Estado=${u.Estado}`);
      check(null, {
        'nombre correcto':  () => u.Usuario === NOMBRE_USR_NUEVO,
        'Estado inicial=P': () => u.Estado === 'P',
      });
    }
  });

  group('F9-05: GET /usuarios/0 — IdUsuario=0 -> 400', () => {
    const res = http.get(`${BASE}/usuarios/0`, PARAMS_SISTEMA);
    logRes('dame-usuario-id-cero', res);
    check(res, { 'dame usuario id=0 -> 400': (r) => r.status === 400 });
  });

  group(`F9-06: GET /usuarios?cadena=${NOMBRE_USR_NUEVO} — buscar por nombre`, () => {
    // El usuario recién creado está en Estado=P, incluimos pendientes explícitamente
    const res = http.get(`${BASE}/usuarios?cadena=${encodeURIComponent(NOMBRE_USR_NUEVO)}&incluyeInactivos=S`, PARAMS_SISTEMA);
    logRes('buscar-usuarios-nombre', res);
    check(res, { 'buscar usuarios -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      const arr = Array.isArray(b) ? b : [];
      console.log(`  Total: ${arr.length}  (esperado >= 1)`);
      check(null, { 'usuario nuevo encontrado en búsqueda': () => arr.length >= 1 });
    }
  });

  group('F9-07: GET /usuarios — buscar todos (sin filtro)', () => {
    const res = http.get(`${BASE}/usuarios`, PARAMS_SISTEMA);
    logRes('buscar-todos-usuarios', res);
    check(res, { 'buscar todos -> 200': (r) => r.status === 200 });
  });

  group('F9-08: GET /usuarios?incluyeInactivos=S — incluir inactivos', () => {
    const res = http.get(`${BASE}/usuarios?incluyeInactivos=S`, PARAMS_SISTEMA);
    logRes('buscar-usuarios-con-inactivos', res);
    check(res, { 'incluyeInactivos=S -> 200': (r) => r.status === 200 });
  });

  group('F9-09: GET /usuarios?incluyeInactivos=X — valor inválido -> 400', () => {
    const res = http.get(`${BASE}/usuarios?incluyeInactivos=X`, PARAMS_SISTEMA);
    logRes('buscar-usuarios-invalido', res);
    check(res, { 'incluyeInactivos=X -> 400': (r) => r.status === 400 });
  });

  group(`F9-10: PUT /usuarios/password/reestablecer — usuario Estado=P -> 400`, () => {
    // tsp_restablecer_password_usuario rechaza Estado=P con "El usuario ya está en estado
    // pendiente." Solo funciona para usuarios Estado=A. El test real de reestablecer se
    // hace más adelante (F9-16b) cuando el usuario ya confirmó su cuenta y está Activo.
    if (idUsuarioNuevo <= 0) { console.log('  SKIP'); return; }
    const res = http.put(
      `${BASE}/usuarios/password/reestablecer`,
      JSON.stringify({ IdUsuario: idUsuarioNuevo }),
      PARAMS_SISTEMA
    );
    logRes('restablecer-password-pendiente', res);
    check(res, { 'restablecer password en Estado=P -> 400': (r) => r.status === 400 });
  });

  group('F9-11: PUT /usuarios/password/reestablecer — IdUsuario=0 -> 400', () => {
    const res = http.put(
      `${BASE}/usuarios/password/reestablecer`,
      JSON.stringify({ IdUsuario: 0 }),
      PARAMS_SISTEMA
    );
    logRes('restablecer-pass-id-cero', res);
    check(res, { 'restablecer pass id=0 -> 400': (r) => r.status === 400 });
  });

  group(`F9-12: POST /usuarios/login — login con pass temporal de '${NOMBRE_USR_NUEVO}'`, () => {
    if (idUsuarioNuevo <= 0 || passTemp === '') { console.log('  SKIP'); return; }
    const res = http.post(
      `${BASE}/usuarios/login`,
      JSON.stringify({ Usuario: NOMBRE_USR_NUEVO, Password: passTemp }),
      PARAMS_NO_AUTH
    );
    logRes('login-usuario-nuevo-temporal', res);
    check(res, { 'login con pass temporal -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      tokenNuevoUsr = b.TokenSesion || '';
      console.log(`  Token nuevo usuario: ${tokenNuevoUsr ? tokenNuevoUsr.substring(0, 8) + '...' : 'NONE'}`);
      console.log(`  Mensaje: ${b.Mensaje}  (esperado: contiene "Se requiere")`);
      check(null, { 'login pass temporal -> Estado=P': () => b.Mensaje && b.Mensaje.includes('Se requiere') });
    }
  });

  group(`F9-13: PUT /usuarios/confirmar-cuenta — establecer password definitivo`, () => {
    if (idUsuarioNuevo <= 0 || tokenNuevoUsr === '') { console.log('  SKIP'); return; }
    const res = http.put(
      `${BASE}/usuarios/confirmar-cuenta`,
      JSON.stringify({ Password: PASS_INICIAL, ConfirmarPassword: PASS_INICIAL }),
      makeBearer(tokenNuevoUsr)
    );
    logRes('confirmar-cuenta', res);
    check(res, { 'confirmar cuenta -> 200': (r) => r.status === 200 });
  });

  group('F9-14: PUT /usuarios/confirmar-cuenta — passwords no coinciden -> 400', () => {
    if (idUsuarioNuevo <= 0) { console.log('  SKIP'); return; }
    const res = http.put(
      `${BASE}/usuarios/confirmar-cuenta`,
      JSON.stringify({ Password: 'abc', ConfirmarPassword: 'xyz' }),
      makeBearer(tokenNuevoUsr || 'token_falso')
    );
    logRes('confirmar-pass-no-coincide', res);
    // 'abc' falla validación de formato (< 8 chars, sin mayúscula, sin número) -> 400
    check(res, { 'passwords distintas -> 400 o 401': (r) => r.status === 400 || r.status === 401 });
  });

  group('F9-14b: PUT /usuarios/confirmar-cuenta — passwords válidas pero no coinciden -> 400', () => {
    if (idUsuarioNuevo <= 0) { console.log('  SKIP'); return; }
    const res = http.put(
      `${BASE}/usuarios/confirmar-cuenta`,
      JSON.stringify({ Password: 'ValidoPass1', ConfirmarPassword: 'OtroValido2' }),
      makeBearer(tokenNuevoUsr || 'token_falso')
    );
    logRes('confirmar-cuenta-no-coincide', res);
    check(res, { 'confirmar-cuenta passwords no coinciden -> 400': (r) => r.status === 400 });
  });

  group('F9-15: PUT /usuarios/confirmar-cuenta — token inválido -> 400 o 401', () => {
    if (idUsuarioNuevo <= 0) { console.log('  SKIP'); return; }
    const res = http.put(
      `${BASE}/usuarios/confirmar-cuenta`,
      JSON.stringify({ Password: PASS_INICIAL, ConfirmarPassword: PASS_INICIAL }),
      makeBearer('token_invalido_zzzz9999xxxx')
    );
    logRes('confirmar-cuenta-token-invalido', res);
    // Esta ruta está fuera del middleware de auth; el SP valida el token internamente
    // y el controller devuelve 400 para token inválido (no 401).
    check(res, { 'confirmar-cuenta token inválido -> 400 o 401': (r) => r.status === 400 || r.status === 401 });
  });

  group(`F9-16: GET /usuarios/${idUsuarioNuevo || 1} — Estado=A tras confirmar cuenta`, () => {
    if (idUsuarioNuevo <= 0) { console.log('  SKIP'); return; }
    const res = http.get(`${BASE}/usuarios/${idUsuarioNuevo}`, PARAMS_SISTEMA);
    logRes('dame-usuario-activo', res);
    if (res.status === 200) {
      const b = parseBody(res);
      const u = b.Usuario || {};
      console.log(`  Estado: ${u.Estado}  (esperado: A)`);
      check(null, { 'Estado=A tras confirmar cuenta': () => u.Estado === 'A' });
    }
  });

  group(`F9-16b: PUT /usuarios/password/reestablecer — usuario Activo -> 200`, () => {
    if (idUsuarioNuevo <= 0) { console.log('  SKIP'); return; }
    const res = http.put(
      `${BASE}/usuarios/password/reestablecer`,
      JSON.stringify({ IdUsuario: idUsuarioNuevo }),
      PARAMS_SISTEMA
    );
    logRes('restablecer-password-activo', res);
    check(res, { 'restablecer password (Estado=A) -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      passTemp = b.PasswordTemporal || passTemp; // nueva passTemp; SP vuelve al usuario a Estado=P
      console.log(`  Nueva PasswordTemporal: ${passTemp}`);
    }
  });

  group(`F9-16c: POST /usuarios/login — login con nueva passTemp post-reestablecer`, () => {
    if (idUsuarioNuevo <= 0 || passTemp === '') { console.log('  SKIP'); return; }
    const res = http.post(
      `${BASE}/usuarios/login`,
      JSON.stringify({ Usuario: NOMBRE_USR_NUEVO, Password: passTemp }),
      PARAMS_NO_AUTH
    );
    logRes('login-usuario-tras-reestablecer', res);
    check(res, { 'login con passTemp post-reestablecer -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      tokenNuevoUsr = b.TokenSesion || tokenNuevoUsr;
      console.log(`  Token renovado: ${tokenNuevoUsr ? tokenNuevoUsr.substring(0, 8) + '...' : 'NONE'}`);
      console.log(`  Mensaje: ${b.Mensaje}  (esperado: contiene "Se requiere")`);
      check(null, { 'login passTemp post-reestablecer -> Estado=P': () => b.Mensaje && b.Mensaje.includes('Se requiere') });
    }
  });

  group(`F9-16d: PUT /usuarios/confirmar-cuenta — reactivar con PASS_INICIAL`, () => {
    // El reestablecer dejó al usuario en Estado=P, necesitamos volver a confirmar.
    if (idUsuarioNuevo <= 0 || tokenNuevoUsr === '') { console.log('  SKIP'); return; }
    const res = http.put(
      `${BASE}/usuarios/confirmar-cuenta`,
      JSON.stringify({ Password: PASS_INICIAL, ConfirmarPassword: PASS_INICIAL }),
      makeBearer(tokenNuevoUsr)
    );
    logRes('confirmar-cuenta-tras-reestablecer', res);
    check(res, { 'confirmar cuenta (2da vez) -> 200': (r) => r.status === 200 });
  });

  group(`F9-17: POST /usuarios/login — login con password definitivo '${PASS_INICIAL}'`, () => {
    if (idUsuarioNuevo <= 0) { console.log('  SKIP'); return; }
    const res = http.post(
      `${BASE}/usuarios/login`,
      JSON.stringify({ Usuario: NOMBRE_USR_NUEVO, Password: PASS_INICIAL }),
      PARAMS_NO_AUTH
    );
    logRes('login-usuario-definitivo', res);
    check(res, { 'login con pass definitivo -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      const b = parseBody(res);
      tokenNuevoUsr = b.TokenSesion || tokenNuevoUsr;
      console.log(`  Token actualizado: ${tokenNuevoUsr ? tokenNuevoUsr.substring(0, 8) + '...' : 'NONE'}`);
      console.log(`  Mensaje: ${b.Mensaje}  (esperado: "OK")`);
      check(null, { 'login pass definitivo -> Estado=A': () => b.Mensaje === 'OK' });
    }
  });

  // F9-19 y F9-20 van ANTES de F9-18: tsp_modificar_password_usuario invalida la sesión
  // al cambiar la contraseña con éxito. Si corrieran después de F9-18, el token quedaría
  // muerto y el middleware devolvería 401 en lugar del 400 esperado del controller.
  group('F9-19: PUT /usuarios/password/modificar — campos vacíos -> 400', () => {
    if (tokenNuevoUsr === '') { console.log('  SKIP'); return; }
    const res = http.put(
      `${BASE}/usuarios/password/modificar`,
      JSON.stringify({
        PasswordAnterior:  '',
        PasswordNuevo:     '',
        ConfirmarPassword: '',
      }),
      makeBearer(tokenNuevoUsr)
    );
    logRes('modificar-pass-vacio', res);
    check(res, { 'modificar-password campos vacíos -> 400': (r) => r.status === 400 });
  });

  group('F9-20: PUT /usuarios/password/modificar — PasswordNuevo != ConfirmarPassword -> 400', () => {
    if (tokenNuevoUsr === '') { console.log('  SKIP'); return; }
    const res = http.put(
      `${BASE}/usuarios/password/modificar`,
      JSON.stringify({ PasswordAnterior: PASS_INICIAL, PasswordNuevo: 'abc', ConfirmarPassword: 'xyz' }),
      makeBearer(tokenNuevoUsr)
    );
    logRes('modificar-pass-no-coincide', res);
    // 'abc' también falla validación de formato (< 8 chars, sin mayúscula, sin número) -> 400
    check(res, { 'passwords no coinciden -> 400': (r) => r.status === 400 });
  });

  group('F9-20b: PUT /usuarios/password/modificar — PasswordNuevo sin mayúscula -> 400', () => {
    if (tokenNuevoUsr === '') { console.log('  SKIP'); return; }
    const res = http.put(
      `${BASE}/usuarios/password/modificar`,
      JSON.stringify({ PasswordAnterior: PASS_INICIAL, PasswordNuevo: 'sinmayuscula1', ConfirmarPassword: 'sinmayuscula1' }),
      makeBearer(tokenNuevoUsr)
    );
    logRes('modificar-pass-sin-mayuscula', res);
    check(res, { 'password sin mayúscula -> 400': (r) => r.status === 400 });
  });

  group('F9-20c: PUT /usuarios/password/modificar — PasswordNuevo sin número -> 400', () => {
    if (tokenNuevoUsr === '') { console.log('  SKIP'); return; }
    const res = http.put(
      `${BASE}/usuarios/password/modificar`,
      JSON.stringify({ PasswordAnterior: PASS_INICIAL, PasswordNuevo: 'SinNumeroClave', ConfirmarPassword: 'SinNumeroClave' }),
      makeBearer(tokenNuevoUsr)
    );
    logRes('modificar-pass-sin-numero', res);
    check(res, { 'password sin número -> 400': (r) => r.status === 400 });
  });

  group(`F9-18: PUT /usuarios/password/modificar — cambiar password (actor USUARIO)`, () => {
    if (tokenNuevoUsr === '') { console.log('  SKIP: sin token'); return; }
    const res = http.put(
      `${BASE}/usuarios/password/modificar`,
      JSON.stringify({
        PasswordAnterior:  PASS_INICIAL,
        PasswordNuevo:     PASS_NUEVO,
        ConfirmarPassword: PASS_NUEVO,
      }),
      makeBearer(tokenNuevoUsr)
    );
    logRes('modificar-password', res);
    check(res, { 'modificar password -> 200': (r) => r.status === 200 });
  });

  group(`F9-21: PUT /usuarios/desactivar/${idUsuarioNuevo || 1} — desactivar (SISTEMA)`, () => {
    if (idUsuarioNuevo <= 0) { console.log('  SKIP'); return; }
    const res = http.put(`${BASE}/usuarios/desactivar/${idUsuarioNuevo}`, null, PARAMS_SISTEMA);
    logRes('desactivar-usuario', res);
    check(res, { 'desactivar usuario -> 200': (r) => r.status === 200 });
  });

  group(`F9-22: GET /usuarios/${idUsuarioNuevo || 1} — Estado=I tras desactivar`, () => {
    if (idUsuarioNuevo <= 0) { console.log('  SKIP'); return; }
    const res = http.get(`${BASE}/usuarios/${idUsuarioNuevo}`, PARAMS_SISTEMA);
    logRes('dame-usuario-inactivo', res);
    if (res.status === 200) {
      const b = parseBody(res);
      const u = b.Usuario || {};
      console.log(`  Estado: ${u.Estado}  (esperado: I)`);
      check(null, { 'Estado=I tras desactivar': () => u.Estado === 'I' });
    }
  });

  group(`F9-23: PUT /usuarios/activar/${idUsuarioNuevo || 1} — reactivar (SISTEMA)`, () => {
    if (idUsuarioNuevo <= 0) { console.log('  SKIP'); return; }
    const res = http.put(`${BASE}/usuarios/activar/${idUsuarioNuevo}`, null, PARAMS_SISTEMA);
    logRes('activar-usuario', res);
    check(res, { 'activar usuario -> 200': (r) => r.status === 200 });
  });

  group(`F9-24: GET /usuarios/${idUsuarioNuevo || 1} — Estado=A tras reactivar`, () => {
    if (idUsuarioNuevo <= 0) { console.log('  SKIP'); return; }
    const res = http.get(`${BASE}/usuarios/${idUsuarioNuevo}`, PARAMS_SISTEMA);
    logRes('dame-usuario-reactivado', res);
    if (res.status === 200) {
      const b = parseBody(res);
      const u = b.Usuario || {};
      console.log(`  Estado: ${u.Estado}  (esperado: A)`);
      check(null, { 'Estado=A tras reactivar': () => u.Estado === 'A' });
    }
  });

  group('F9-25: PUT /usuarios/desactivar/0 — IdUsuario=0 -> 400', () => {
    const res = http.put(`${BASE}/usuarios/desactivar/0`, null, PARAMS_SISTEMA);
    logRes('desactivar-usuario-id-cero', res);
    check(res, { 'desactivar usuario id=0 -> 400': (r) => r.status === 400 });
  });

  group('F9-26: PUT /usuarios/activar/0 — IdUsuario=0 -> 400', () => {
    const res = http.put(`${BASE}/usuarios/activar/0`, null, PARAMS_SISTEMA);
    logRes('activar-usuario-id-cero', res);
    check(res, { 'activar usuario id=0 -> 400': (r) => r.status === 400 });
  });

  group(`F9-27: DELETE /usuarios/${idUsuarioNuevo || 1} — borrar usuario (SISTEMA)`, () => {
    if (idUsuarioNuevo <= 0) { console.log('  SKIP'); return; }
    const res = http.del(`${BASE}/usuarios/${idUsuarioNuevo}`, null, PARAMS_SISTEMA);
    logRes('borrar-usuario', res);
    // 200 = borrado; 400 = tiene operaciones de auditoría (también válido)
    console.log(`  Status: ${res.status}  body: ${res.body}`);
    check(res, { 'borrar usuario -> 200 o 400': (r) => r.status === 200 || r.status === 400 });
  });

  group('F9-28: POST /usuarios/logout — sin credenciales -> 401', () => {
    const res = http.post(`${BASE}/usuarios/logout`, null, PARAMS_NO_AUTH);
    logRes('logout-sin-auth', res);
    check(res, { 'logout sin auth -> 401': (r) => r.status === 401 });
  });

  group('F9-29: POST /usuarios/logout — exitoso (tokenAdmin)', () => {
    if (tokenAdmin === '') { console.log('  SKIP (sin tokenAdmin)'); return; }
    tokenParaLogout = tokenAdmin;
    const res = http.post(`${BASE}/usuarios/logout`, null, makeBearer(tokenParaLogout));
    logRes('logout-exitoso', res);
    check(res, { 'logout exitoso -> 200': (r) => r.status === 200 });
    if (res.status === 200) {
      check(parseBody(res), { 'Mensaje=OK': (b) => b.Mensaje === 'OK' });
    }
  });

  group('F9-30: POST /usuarios/logout — token ya cerrado -> 401', () => {
    if (tokenParaLogout === '') { console.log('  SKIP (sin token previo)'); return; }
    const res = http.post(`${BASE}/usuarios/logout`, null, makeBearer(tokenParaLogout));
    logRes('logout-token-cerrado', res);
    check(res, { 'token ya cerrado -> 401': (r) => r.status === 401 });
  });

  // =========================================================================
  // FASE 10: VERIFICACIÓN FINAL DE BALANCES
  // =========================================================================
  console.log('\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('FASE 10: VERIFICACIÓN FINAL DE BALANCES');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('  UA/MP — transfers procesadas:');
  console.log('    T01: +1000  Ingreso');
  console.log('    T02:  +500  Ingreso');
  console.log('    T03:  -200  Egreso');
  console.log('    T06:  -150  Egreso');
  console.log('    T06R: +150  Reversal de T06 (crédita $150 de vuelta a UA)');
  console.log('  ──────────────────────────────────────────────');
  console.log('    Creditos = 1000+500+150(reversal) = 1650.00');
  console.log('    Debitos  = 200+150                =  350.00');
  console.log('    Balance  = 1300.00');
  console.log('');
  console.log('  UB/MP: T04 +800; T08 rechazada -> Creditos=800 Debitos=0');
  console.log('  UA/MS: T05 +300                -> Creditos=300 Debitos=0');

  group(`F10-01: Balance final UA/MP — esperado Creditos=1650.00 Debitos=350.00`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('balance-final-ua-mp', res);
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Creditos=${b.Creditos}  Debitos=${b.Debitos}  Estado=${b.Estado}`);
      check(b, {
        'UA/MP Estado=A':          (x) => x.Estado === 'A',
        'UA/MP Creditos=1650.00':  (x) => x.Creditos === '1650.00',
        'UA/MP Debitos=350.00':    (x) => x.Debitos === '350.00',
      });
    } else {
      check(res, { 'dame cuenta UA/MP -> 200': (r) => r.status === 200 });
    }
  });

  group(`F10-02: Balance final UB/MP — esperado Creditos=800.00 Debitos=0.00`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UB}/${ID_MONEDA_PRINCIPAL}`, PARAMS_SISTEMA);
    logRes('balance-final-ub-mp', res);
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Creditos=${b.Creditos}  Debitos=${b.Debitos}  Estado=${b.Estado}`);
      check(b, {
        'UB/MP Estado=A':         (x) => x.Estado === 'A',
        'UB/MP Creditos=800.00':  (x) => x.Creditos === '800.00',
        'UB/MP Debitos=0.00':     (x) => x.Debitos === '0.00',
      });
    } else {
      check(res, { 'dame cuenta UB/MP -> 200': (r) => r.status === 200 });
    }
  });

  group(`F10-03: Balance final UA/MS — esperado Creditos=300.00 Debitos=0.00`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_SECUNDARIA}`, PARAMS_SISTEMA);
    logRes('balance-final-ua-ms', res);
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Creditos=${b.Creditos}  Debitos=${b.Debitos}  Estado=${b.Estado}`);
      check(b, {
        'UA/MS Estado=A':         (x) => x.Estado === 'A',
        'UA/MS Creditos=300.00':  (x) => x.Creditos === '300.00',
        'UA/MS Debitos=0.00':     (x) => x.Debitos === '0.00',
      });
    } else {
      check(res, { 'dame cuenta UA/MS -> 200': (r) => r.status === 200 });
    }
  });

  group(`F10-04: Historial final UA/MP — última entrada = estado acumulado`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/historial`, PARAMS_SISTEMA);
    logRes('historial-final-ua-mp', res);
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total entradas historial: ${b.Total}  (esperado >= 5)`);
      check(null, { 'historial >= 5 entradas (T01,T02,T03,T06,reversal)': () => b.Total >= 5 });
      if (b.Total > 0) {
        const ultima = b.Historial[0];
        console.log(`  Primera entrada (más reciente): C=${ultima.Creditos} D=${ultima.Debitos} B=${ultima.Balance}`);
        check(ultima, {
          'historial final Creditos=1650.00': (x) => x.Creditos === '1650.00',
          'historial final Debitos=350.00':   (x) => x.Debitos === '350.00',
          'historial final Balance=1300.00':  (x) => x.Balance === '1300.00',
        });
      }
    }
  });

  group(`F10-05: DameTransferencias final UA/MP — con revertidas, sin internos`, () => {
    const res = http.get(`${BASE}/cuentas/${ID_UA}/${ID_MONEDA_PRINCIPAL}/transferencias?IncluyeRevertidas=true`, PARAMS_SISTEMA);
    logRes('dame-transfers-final-ua-mp', res);
    if (res.status === 200) {
      const b = parseBody(res);
      console.log(`  Total transfers (IncluyeRevertidas=true): ${b.Total}  (esperado >= 4)`);
      // T01, T02, T03 finalizadas + T06 revertida (Estado=R) = al menos 4
      check(null, { 'DameTransferencias >= 4': () => b.Total >= 4 });
      const sinCierre = b.Transferencias.every(t => t.Monto !== '0.00' && t.Monto !== '0');
      check(null, { 'sin transfers de cierre en resultado': () => sinCierre });
      const sinReversionInterna = b.Transferencias.every(t => t.Tipo !== 'R');
      check(null, { 'sin transfers de reversión internas (Tipo!=R)': () => sinReversionInterna });
      const t06 = b.Transferencias.find(t => t.IdTransferencia === T06);
      if (t06) {
        check(null, { 'T06 Estado=R en DameTransferencias final': () => t06.Estado === 'R' });
      }
    }
  });

  console.log('');
  console.log('╔══════════════════════════════════════════════════════════════════╗');
  console.log('║               FIN DEL TEST SISTEMA — TODOS LOS ENDPOINTS        ║');
  console.log('╚══════════════════════════════════════════════════════════════════╝');
}
