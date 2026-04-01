# Microservicio de Transacciones Financieras (MSTF)

## Descripción

Este repositorio contiene el código fuente correspondiente al Trabajo de Graduación de la carrera de Ingeniería en Computación de la Universidad Nacional de Tucumán (UNT) del alumno Llobeta, Bautista José.

El proyecto consiste en un microservicio diseñado para procesar transferencias financieras de alto volumen. La arquitectura está construida para garantizar inmutabilidad, consistencia y velocidad transaccional mediante el uso de una base de datos no relacional especializada, apoyada por una base de datos relacional para la gestión administrativa y un broker de mensajería para el procesamiento asíncrono en lotes. Incluye además un frontend administrativo para la visualización y control del sistema.

## Tecnologías

- **Backend:** Go (Echo framework)
- **Frontend:** Vue.js + Vite
- **Base de Datos Transaccional:** TigerBeetle
- **Base de Datos Administrativa:** MySQL 8.0
- **Broker de Mensajería:** Kafka (Modo KRaft)
- **Infraestructura:** Docker, Docker Compose
- **Testing:** Go, k6

## Prerrequisitos y Configuración

Para ejecutar el proyecto, es necesario contar con Docker instalado en el sistema.

Antes de iniciar, se debe crear un archivo `.env` en la raíz del repositorio. A continuación se presenta una configuración de ejemplo:

```env
# Backend
PORT=
WEBHOOK_URL=

# MySQL
MYSQL_HOST=
MYSQL_PORT=
MYSQL_ROOT_PASSWORD=
MYSQL_USER=
MYSQL_PASSWORD=
MYSQL_DATABASE=

# Kafka
KAFKA_CLUSTER_ID=
KAFKA_BROKERS=
KAFKA_TOPIC_TRANSFERS=
KAFKA_GROUP_ID=

# TigerBeetle (IP fija asignada en la red de Docker Compose)
TB_ADDRESSES=   # IP:PUERTO
```

Aclaración: se puede usar una variable de entorno VITE_DEMO_MODE para que el frontend se levante con opciones adicionales a sus capacidades, para facilitar las pruebas (permitir creación y borrado de monedas, creación de cuentas y transferencias, etc). Para ello se debe poner en true dicha variable. Por defecto se encuentra deshabilitada.

## Ejecución

El sistema completo (bases de datos, broker de mensajería, backend y frontend) está contenerizado y orquestado. Para levantarlo en su totalidad, basta con ejecutar el siguiente comando desde la raíz del proyecto:

```bash
docker compose --env-file .env up -d --build
```

Una vez finalizado el proceso de construcción e inicialización, los servicios estarán disponibles en los siguientes puertos locales:

- **API REST (Backend):** localhost:{PORT}
- **Interfaz Administrativa (Frontend):** localhost:5173

## Desarrollo

Para agilizar el desarrollo sin necesidad de reconstruir los contenedores repetidamente, los servicios pueden ejecutarse de forma aislada.

### 1. Infraestructura Base

Levantar únicamente las bases de datos y Kafka mediante Docker Compose:

```bash
docker compose up -d mysql kafka tigerbeetle kafka-init
```

### 2. Backend (Go)

Se utiliza la herramienta `air` para recarga automática. Posicionándose en el directorio `/mstf`:

```bash
air
```

_(Nota: Si el backend se ejecuta fuera de la red de Docker, las variables `MYSQL_HOST` y `KAFKA_BROKERS` en el archivo `.env` local deben apuntar a `localhost`)._

### 3. Frontend (Vue)

Posicionándose en el directorio `/frontend`:

```bash
npm install
npm run dev
```

### 4. Ejecución nativa de TigerBeetle (Opcional si se quiere tener fuera del contenedor - Linux/macOS)

Si se requiere ejecutar la base de datos transaccional de forma local sin Docker, se proveen los comandos oficiales de inicialización:

```bash
# Linux
curl -Lo tigerbeetle.zip https://linux.tigerbeetle.com && unzip tigerbeetle.zip
./tigerbeetle version

# macOS
curl -Lo tigerbeetle.zip https://mac.tigerbeetle.com && unzip tigerbeetle.zip
./tigerbeetle version

# Formatear
./tigerbeetle format --cluster=0 --replica=0 --replica-count=1 --development ./0_0.tigerbeetle

# Iniciar
./tigerbeetle start --addresses=3000 --development ./0_0.tigerbeetle
```

Para reiniciar de cero, basta con borrar el archivo 0_0.tigerbeetle y ejecutar nuevamente los comandos de formateo e iniciación.

## Testing

El repositorio cuenta con dos pruebas para validar la integridad y el rendimiento del microservicio.

### 1. Pruebas de Sistema (`testSistema.js`)

Script desarrollado para la herramienta `k6`. Ejecuta llamados HTTP que cubren la totalidad de los endpoints del sistema en la mayoría de escenarios posibles. Valida reglas de negocio, manejo de errores, autenticación, etc. Al finalizar, el test habrá poblado la moneda configurada con cuentas y transferencias. Por lo que si se desea correrlo múltiples veces sin reiniciar las bases de datos, se pueden modificar el id de la moneda a crear.

```bash
k6 run testSistema.js
```

### 2. Prueba de Estrés (`stress.go`)

Script desarrollado en Go para evaluar el rendimiento máximo de procesamiento asíncrono y TPS (Transacciones por Segundo) del sistema. El test crea N monedas y M cuentas en cada moneda. Luego espera a que manualmente se apague el contenedor mstf-app y puebla de K transferencias la cola de kafka. Una vez llenada esta cola, se espera a levantar manualmente el contenedor mencionado y cuando éste empieza a consumir de la cola, se empieza el timer que mide el tiempo de procesamiento.

```bash
go run stress.go --transferencias=100000
```

**Importante:** El microservicio notifica la finalización de los lotes de transferencia mediante Webhooks. Para que el test de estrés funcione y mida los tiempos correctamente, la variable `WEBHOOK_URL` del archivo `.env` del backend debe coincidir con la dirección IP y el puerto donde se ejecuta este script.

### 3. Test SQL (`testSPs.sql`)

Este simplemente es un test que ejecuta llamados a todos los SPs de la base de datos MySQL. Sirve para validar el correcto funcionamiento de los mismos.

## Documentación

La especificación técnica de la API REST se encuentra en el archivo `swagger.yaml` ubicado en la raíz del repositorio. Este archivo describe en detalle los endpoints, métodos permitidos, parámetros de entrada y esquemas de respuesta. Para ser visualizado se recomienda utilizar una extensión de Swagger para navegadores.
