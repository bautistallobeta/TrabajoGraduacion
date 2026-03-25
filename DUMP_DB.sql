CREATE DATABASE  IF NOT EXISTS `mstf` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;
USE `mstf`;
-- MySQL dump 10.13  Distrib 8.0.41, for Win64 (x86_64)
--
-- Host: 127.0.0.1    Database: mstf
-- ------------------------------------------------------
-- Server version	8.0.41

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `Monedas`
--

DROP TABLE IF EXISTS `Monedas`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `Monedas` (
  `IdMoneda` int NOT NULL COMMENT 'PK de la tabla Monedas. Referencia el IdMoneda del sistema cliente.',
  `IdCuentaEmpresa` varchar(50) DEFAULT NULL,
  `FechaAlta` datetime NOT NULL COMMENT 'Fecha en que se creó la Moneda.',
  `Estado` char(1) NOT NULL COMMENT 'Estado de la Moneda: A (Activo) - I (Inactivo) - P (Pendiente)',
  PRIMARY KEY (`IdMoneda`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;



--
-- Table structure for table `Operaciones`
--

DROP TABLE IF EXISTS `Operaciones`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `Operaciones` (
  `IdOperacion` int NOT NULL AUTO_INCREMENT COMMENT 'PK de la tabla Operaciones.',
  `IdUsuario` int DEFAULT NULL COMMENT 'FK a la tabla Usuarios. NULL cuando la operación la realiza el sistema.',
  `TipoOperacion` char(2) NOT NULL COMMENT 'Tipo de operación que se audita: CM (creación de moneda) - AM (activación de moneda) - DM (desactivación de moneda) - BM (borrado de moneda) - MP (modificación de parámetro) - CU (creación de usuario) - AU (activación de usuario) - DU (desactivación de usuario) - BU (borrado de usuario)',
  `FechaOperacion` datetime NOT NULL,
  `Detalles` json NOT NULL,
  PRIMARY KEY (`IdOperacion`),
  KEY `Ref22` (`IdUsuario`),
  CONSTRAINT `RefUsuarios2` FOREIGN KEY (`IdUsuario`) REFERENCES `Usuarios` (`IdUsuario`)
) ENGINE=InnoDB AUTO_INCREMENT=444 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Tabla de auditoría de operaciones administrativas realizadas en el MSTF.';
/*!40101 SET character_set_client = @saved_cs_client */;


--
-- Table structure for table `Parametros`
--

DROP TABLE IF EXISTS `Parametros`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `Parametros` (
  `Parametro` varchar(50) NOT NULL COMMENT 'Nombre del Parametro.. Es único y PK.',
  `Valor` text NOT NULL COMMENT 'Valor del Parámetro.',
  `Descripcion` varchar(255) DEFAULT NULL COMMENT 'Descripción del parámetro. Opcional.',
  `EsModificable` char(1) NOT NULL COMMENT 'Define si un parámetro es del sistema o si puede ser modificado por un usuario administrativo. S (Si es modificable) - N (No es modificable)',
  PRIMARY KEY (`Parametro`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Tabla que almacena los Parámetros que se puedan definir con su valor.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `Parametros`
--

LOCK TABLES `Parametros` WRITE;
/*!40000 ALTER TABLE `Parametros` DISABLE KEYS */;
INSERT INTO `Parametros` VALUES ('APIKEY_SISTEMA','CAMBIAR_ESTE_VALOR','API Key del sistema cliente externo','S'),('KAFKABATCHSIZE','8189','Cantidad máxima de transferencias que se procesan en un lote desde Kafka','N'),('KAFKABATCHTIMEOUTMS','500','Tiempo máximo en milisegundos para armar un lote de transferencias desde Kafka antes de procesarlo','N'),('LIMITEBUSCARCUENTAS','500','Cantidad máxima de cuentas a devolver en una búsqueda cuando no se especifica límite en la consulta','S'),('LIMITEBUSCARTRANSFERENCIAS','500','Cantidad máxima de transferencias a devolver en una búsqueda cuando no se especifica límite en la consulta','S'),('LIMITEHISTORIALBALANCE','500','Cantidad máxima de entradas a devolver en el historial de balances de una cuenta cuando no se especifica\n   límite en la consulta','S'),('LIMITEMAXIMOBUSCARCUENTAS','500','Cantidad máxima absoluta de cuentas que puede solicitar un cliente en una búsqueda','S'),('MONTOMAXTRANSFER','100000','Monto máximo permitido para transferencias','S'),('MONTOMINTRANSFER','100','Monto mínimo permitido para transferencias','S'),('RETRYBACKOFFMAXSEG','20','Tiempo máximo en segundos del backoff exponencial al reintentar un lote fallido','N'),('version_api','1.0.0','Versión actual de la API','N');
/*!40000 ALTER TABLE `Parametros` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `Usuarios`
--

DROP TABLE IF EXISTS `Usuarios`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `Usuarios` (
  `IdUsuario` int NOT NULL AUTO_INCREMENT COMMENT 'PK de la tabla Usuarios.',
  `Usuario` varchar(30) NOT NULL COMMENT 'Nombre de usuario del mismo. Es único.',
  `Password` char(32) NOT NULL COMMENT 'Clave encriptada en MD5.',
  `TokenSesion` char(32) NOT NULL COMMENT 'Token de sesión del cliente. Generado aleatoriamente y hasheado en MD5.',
  `FechaAlta` datetime NOT NULL COMMENT 'Fecha de creación del Usuario.',
  `Estado` char(1) NOT NULL COMMENT 'Estado del Usuario: A (Activo) - I (Inactivo)',
  `Rol` char(1) NOT NULL DEFAULT 'O',
  PRIMARY KEY (`IdUsuario`),
  UNIQUE KEY `UI_Usuario` (`Usuario`),
  UNIQUE KEY `UI_TokenSesion` (`TokenSesion`)
) ENGINE=InnoDB AUTO_INCREMENT=28 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Tabla que almacena los Usuarios Administradores que gestionan aspectos del MSTF mediante el sitio administrativo.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `Usuarios`
--

LOCK TABLES `Usuarios` WRITE;
/*!40000 ALTER TABLE `Usuarios` DISABLE KEYS */;
INSERT INTO `Usuarios` VALUES (1,'admin','e64b78fc3bc91bcbc7dc232ba8ec59e0','257f899c88eb27dd7b281f6b8ba29872','2026-02-13 21:29:32','A','A');
/*!40000 ALTER TABLE `Usuarios` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;


--
-- Dumping routines for database 'mstf'
--
/*!50003 DROP FUNCTION IF EXISTS `f_valida_usuario` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` FUNCTION `f_valida_usuario`(pTokenSesion char(32)) RETURNS int
    READS SQL DATA
    DETERMINISTIC
BEGIN
	/*
    Valida que exista el token del usuario y que éste se encuentre activo. 
    En caso que no se valide, devuelve 0. Si es válido, devuelve el IdUsuario.
    */
    RETURN (SELECT COALESCE(MAX(IdUsuario),0) FROM Usuarios WHERE TokenSesion = pTokenSesion AND Estado = 'A');
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_activar_moneda` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_activar_moneda`(
    pCredencial VARCHAR(255),
    pActor CHAR(10),
    pIdMoneda INT
)
SALIR: BEGIN
    /*
    Activa una moneda en Estado Inactivo o Pendiente.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuario INT DEFAULT NULL;
    DECLARE pEstado CHAR(1);
    DECLARE pLog VARCHAR(100);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
    END;
    IF pActor = 'USUARIO' THEN
        SET pIdUsuario = f_valida_usuario(pCredencial);
        IF pIdUsuario = 0 THEN
            SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
            LEAVE SALIR;
        END IF;
    END IF;
    
    SELECT Estado INTO pEstado
    FROM Monedas
    WHERE IdMoneda = pIdMoneda;

    IF pEstado IS NULL THEN
        SELECT 'La moneda no existe.' Mensaje;
        LEAVE SALIR;
    END IF;
    IF pEstado = 'A' THEN
		SELECT 'La moneda ya está en estado Activo.' Mensaje;
        LEAVE SALIR;
    END IF;
    IF pEstado NOT IN ('P', 'I') THEN
        SELECT 'La moneda no está en estado Inactivo ni Pendiente.' Mensaje;
        LEAVE SALIR;
    END IF;

    UPDATE Monedas
    SET Estado = 'A'
    WHERE IdMoneda = pIdMoneda;
    SET pLog = IF(pEstado = 'P', 'Activación luego de creación de moneda en TigerBeetle', 'Activación manual de moneda');
    INSERT INTO Operaciones (IdUsuario, TipoOperacion, FechaOperacion, Detalles)
    VALUES (
        pIdUsuario,
        'AM',
        NOW(),
        JSON_OBJECT('IdMoneda', pIdMoneda, 'Log', pLog)
    );

    SELECT 'OK' Mensaje;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_activar_usuario` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_activar_usuario`(pIdUsuario int, pCredencial VARCHAR(255), pActor
  CHAR(10))
SALIR: BEGIN
      /*
      Permite cambiar el estado de un usuario a A: Activo siempre y cuando esté inactivo.
      Devuelve OK o el mensaje de error en Mensaje.
      Mensaje varchar(100)
      */
      DECLARE pEstado char(1);
      DECLARE pIdUsuarioActor INT DEFAULT NULL;
      -- Manejo de error en la transacción
      DECLARE EXIT HANDLER FOR SQLEXCEPTION
      BEGIN
          SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
      END;
      -- Resuelve identidad del actor
      IF pActor = 'USUARIO' THEN
          SET pIdUsuarioActor = f_valida_usuario(pCredencial);
          IF pIdUsuarioActor = 0 THEN
              SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
              LEAVE SALIR;
          END IF;
      END IF;
      -- Controla que el actor tenga rol Administrador
      IF pActor = 'USUARIO' AND (SELECT Rol FROM Usuarios WHERE IdUsuario = pIdUsuarioActor) != 'A' THEN
          SELECT 'No tienes permisos para realizar esta acción.' Mensaje;
          LEAVE SALIR;
      END IF;
      -- Controla que el actor no se modifique a sí mismo
      IF pActor = 'USUARIO' AND pIdUsuarioActor = pIdUsuario THEN
          SELECT 'No puedes realizar esta acción sobre tu propia cuenta.' Mensaje;
          LEAVE SALIR;
      END IF;
      -- Controla parámetros
      SET pEstado = (SELECT Estado FROM Usuarios WHERE IdUsuario = pIdUsuario);
      IF pEstado IS NULL THEN
          SELECT 'El usuario no existe.' Mensaje;
          LEAVE SALIR;
      END IF;
      IF pEstado = 'A' THEN
          SELECT 'El usuario ya está activo.' Mensaje;
          LEAVE SALIR;
      END IF;
      IF pEstado = 'P' THEN
          SELECT 'El usuario está pendiente y debe confirmar su cuenta.' Mensaje;
          LEAVE SALIR;
      END IF;
      -- Activa
      UPDATE  Usuarios
      SET     Estado = 'A'
      WHERE   IdUsuario = pIdUsuario;
      -- Audita
      INSERT INTO Operaciones (IdUsuario, TipoOperacion, FechaOperacion, Detalles)
      VALUES (pIdUsuarioActor, 'AU', NOW(), JSON_OBJECT('IdUsuario', pIdUsuario));
      SELECT 'OK' Mensaje;
  END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_autenticar_actor` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_autenticar_actor`(pCredencial VARCHAR(255), pActor CHAR(10))
BEGIN
    /*
    Valida las credenciales de un actor (USUARIO o SISTEMA).
    Para USUARIO: verifica que el token de sesión sea válido y el usuario esté activo.
    Para SISTEMA: verifica que la API key coincida con el parámetro APIKEY_SISTEMA.
    Devuelve 'OK' o un mensaje de error.
    */
    IF pActor = 'USUARIO' THEN
        IF f_valida_usuario(pCredencial) = 0 THEN
            SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        ELSE
            SELECT 'OK' Mensaje;
        END IF;
    ELSEIF pActor = 'SISTEMA' THEN
        IF NOT EXISTS (
            SELECT 1 FROM Parametros
            WHERE Parametro = 'APIKEY_SISTEMA' AND Valor = pCredencial
        ) THEN
            SELECT 'API Key inválida.' Mensaje;
        ELSE
            SELECT 'OK' Mensaje;
        END IF;
    ELSE
        SELECT 'Actor inválido.' Mensaje;
    END IF;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_borrar_moneda` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_borrar_moneda`(
    pCredencial VARCHAR(255),
    pActor CHAR(10),
    pIdMoneda INT
)
SALIR: BEGIN
    /*
    Borra una moneda únicamente si está en estado Inactivo o Pendiente.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuario INT DEFAULT NULL;
    DECLARE pEstado CHAR(1);
    DECLARE pLog VARCHAR(100);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
    END;

    IF pActor = 'USUARIO' THEN
        SET pIdUsuario = f_valida_usuario(pCredencial);
        IF pIdUsuario = 0 THEN
            SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
            LEAVE SALIR;
        END IF;
    END IF;

    SELECT Estado INTO pEstado
    FROM Monedas
    WHERE IdMoneda = pIdMoneda;

    IF pEstado IS NULL THEN
        SELECT 'La moneda no existe.' Mensaje;
        LEAVE SALIR;
    END IF;

    IF pEstado NOT IN ('P', 'I') THEN
        SELECT 'Solo se pueden borrar monedas en estado Inactivo o Pendiente.' Mensaje;
        LEAVE SALIR;
    END IF;

    DELETE FROM Monedas WHERE IdMoneda = pIdMoneda;
	SET pLog = IF(pEstado = 'P', 'Rollback por error en creación de moneda', 'Borrado manual de moneda');
    INSERT INTO Operaciones (IdUsuario, TipoOperacion, FechaOperacion, Detalles)
    VALUES (
        pIdUsuario,
        'BM',
        NOW(),
        JSON_OBJECT('IdMoneda', pIdMoneda, 'Log', pLog)
    );

    SELECT 'OK' Mensaje;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_borrar_usuario` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_borrar_usuario`(pIdUsuario int, pCredencial VARCHAR(255), pActor
  CHAR(10))
SALIR: BEGIN
      /*
      Permite eliminar un usuario siempre y cuando no tenga registros en Operaciones.
      Devuelve OK o el mensaje de error en Mensaje.
      Mensaje varchar(100)
      */
      DECLARE pIdUsuarioActor INT DEFAULT NULL;
      DECLARE pUsuarioBorrado VARCHAR(30);
      -- Manejo de error en la transacción
      DECLARE EXIT HANDLER FOR SQLEXCEPTION
      BEGIN
          SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
      END;
      -- Resuelve identidad del actor
      IF pActor = 'USUARIO' THEN
          SET pIdUsuarioActor = f_valida_usuario(pCredencial);
          IF pIdUsuarioActor = 0 THEN
              SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
              LEAVE SALIR;
          END IF;
      END IF;
      -- Controla que el actor tenga rol Administrador
      IF pActor = 'USUARIO' AND (SELECT Rol FROM Usuarios WHERE IdUsuario = pIdUsuarioActor) != 'A' THEN
          SELECT 'No tienes permisos para realizar esta acción.' Mensaje;
          LEAVE SALIR;
      END IF;
      -- Controla existencia y obtiene nombre (para auditoría)
      SET pUsuarioBorrado = (SELECT Usuario FROM Usuarios WHERE IdUsuario = pIdUsuario);
      IF pUsuarioBorrado IS NULL THEN
          SELECT 'El usuario no existe.' Mensaje;
          LEAVE SALIR;
      END IF;
      IF (SELECT Estado FROM Usuarios WHERE IdUsuario = pIdUsuario) = 'A' THEN
          SELECT 'No se puede borrar un usuario Activo.' Mensaje;
          LEAVE SALIR;
      END IF;
      -- Controla que el actor no se elimine a sí mismo
      IF pActor = 'USUARIO' AND pIdUsuarioActor = pIdUsuario THEN
          SELECT 'No puedes realizar esta acción sobre tu propia cuenta.' Mensaje;
          LEAVE SALIR;
      END IF;
      -- Controla que no tenga registros en auditoría
      IF EXISTS(SELECT IdOperacion FROM Operaciones WHERE IdUsuario = pIdUsuario) THEN
          SELECT 'No se puede eliminar el usuario porque tiene operaciones registradas en auditoría.' Mensaje;
          LEAVE SALIR;
      END IF;
      -- Audita antes de eliminar
      INSERT INTO Operaciones (IdUsuario, TipoOperacion, FechaOperacion, Detalles)
      VALUES (pIdUsuarioActor, 'BU', NOW(), JSON_OBJECT('IdUsuario', pIdUsuario, 'Usuario', pUsuarioBorrado));
      -- Elimina
      DELETE FROM Usuarios WHERE IdUsuario = pIdUsuario;
      SELECT 'OK' Mensaje;
  END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_buscar_parametros` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_buscar_parametros`(pCadena varchar(50), pSoloModificables char(1))
SALIR: BEGIN
	/*
    Permite buscar los parámetros del sistema según su nombre. Si pSoloModificables es 'S', muestra solo los
    modificables desde el sitio administrativo. Ordena por nombre de parámetro.
    */

    SET SESSION TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
    
    SELECT		Parametro, Valor, Descripcion, EsModificable
    FROM		Parametros
    WHERE		(pSoloModificables = 'N' OR (EsModificable = 'S' AND pSoloModificables = 'S'))
    ORDER BY	Parametro;
    
	SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_buscar_usuarios` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_buscar_usuarios`(pCadena varchar(50), pIncluyeInactivosPendientes char(1))
SALIR: BEGIN
      /*
      Permite listar todos los usuarios que cumplan con la condición de búsqueda: la cadena debe estar
      contenida en el nombre de usuario. Puede o no incluir los usuarios inactivos y pendientes
      según pIncluyeInactivosPendientes (S: Si - N: No). Ordena por nombre de usuario.
      */

      SET pCadena = COALESCE(pCadena, '');

      SET SESSION TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;

      SELECT      IdUsuario, Usuario, FechaAlta, Estado, Rol
      FROM        Usuarios
      WHERE       (Usuario LIKE CONCAT('%', TRIM(pCadena), '%')) AND
                  (pIncluyeInactivosPendientes = 'S' OR (Estado = 'A' AND pIncluyeInactivosPendientes = 'N'))
      ORDER BY    Usuario;

      SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;
  END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_confirmar_cuenta_usuario` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_confirmar_cuenta_usuario`(pTokenSesion char(32), pPassword varchar(50), pConfirmarPassword
  varchar(50))
SALIR: BEGIN
      DECLARE pIdUsuario INT;
      DECLARE pEstado CHAR(1);

      DECLARE EXIT HANDLER FOR SQLEXCEPTION
      BEGIN
          SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
      END;

      -- Verifica sesión (acepta estado P o A)
      SELECT  IdUsuario, Estado
      INTO    pIdUsuario, pEstado
      FROM    Usuarios
      WHERE   TokenSesion = pTokenSesion AND Estado IN ('A', 'P');

      IF pIdUsuario IS NULL THEN
          SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
          LEAVE SALIR;
      END IF;

      -- Solo usuarios Pendientes pueden usar este SP
      IF pEstado != 'P' THEN
          SELECT 'El usuario ya está activo. Use modificar contraseña.' Mensaje;
          LEAVE SALIR;
      END IF;

      IF (pPassword IS NULL OR pPassword = '') THEN
          SELECT 'La contraseña es obligatoria.' Mensaje;
          LEAVE SALIR;
      END IF;
      IF (pConfirmarPassword IS NULL OR pConfirmarPassword = '') THEN
          SELECT 'La confirmación de contraseña es obligatoria.' Mensaje;
          LEAVE SALIR;
      END IF;
      IF (CHAR_LENGTH(pPassword) < 6) THEN
          SELECT 'La contraseña debe tener al menos 6 caracteres.' Mensaje;
          LEAVE SALIR;
      END IF;
      IF (pPassword NOT REGEXP '[A-Za-z]') THEN
          SELECT 'La contraseña debe incluir al menos una letra.' Mensaje;
          LEAVE SALIR;
      END IF;
      IF (pPassword NOT REGEXP '[0-9]') THEN
          SELECT 'La contraseña debe incluir al menos un número.' Mensaje;
          LEAVE SALIR;
      END IF;
      IF (pPassword != pConfirmarPassword) THEN
          SELECT 'La contraseña no coincide con su confirmación.' Mensaje;
          LEAVE SALIR;
      END IF;

      -- Actualiza contraseña, activa usuario y regenera token
      UPDATE  Usuarios
      SET     `Password` = pPassword,
              Estado = 'A',
              TokenSesion = md5(CONCAT(pIdUsuario, UNIX_TIMESTAMP(), RAND()))
      WHERE   IdUsuario = pIdUsuario;

      SELECT 'OK' Mensaje;
  END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_crear_moneda` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_crear_moneda`(
    pCredencial VARCHAR(255),
    pActor CHAR(10),
    pIdMoneda INT,
    pIdCuentaEmpresa VARCHAR(50)
)
SALIR: BEGIN
    /*
    Crea una moneda en estado P: Pendiente.
    Si existe la moneda en estado P (aún pendiente de finalizar el proceso de creacion), retorna OK.
    Devuelve OK o mensaje de error.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuario INT DEFAULT NULL;
    DECLARE pEstado CHAR(1);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
    END;

    IF pActor = 'USUARIO' THEN
        SET pIdUsuario = f_valida_usuario(pCredencial);
        IF pIdUsuario = 0 THEN
            SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
            LEAVE SALIR;
        END IF;
    END IF;

    IF pIdMoneda IS NULL OR pIdMoneda <= 0 THEN
        SELECT 'El Id de moneda es obligatorio.' Mensaje;
        LEAVE SALIR;
    END IF;
    
    IF pIdCuentaEmpresa IS NULL OR TRIM(pIdCuentaEmpresa) = '' THEN 
		SELECT 'El IdCuentaEmpresa de la moneda es obligatorio.' Mensaje;
        LEAVE SALIR;
    END IF;
    
	SELECT Estado INTO pEstado FROM Monedas WHERE IdMoneda = pIdMoneda;
    
    IF pEstado = 'A' THEN
        SELECT 'La moneda ya existe.' Mensaje;
        LEAVE SALIR;
    END IF;
    
    IF pEstado IS NULL THEN
		INSERT INTO Monedas (IdMoneda, IdCuentaEmpresa, Estado, FechaAlta)
		VALUES (pIdMoneda, pIdCuentaEmpresa, 'P', NOW());
	
		INSERT INTO Operaciones (IdUsuario, TipoOperacion, FechaOperacion, Detalles)
		VALUES (
			pIdUsuario,
			'CM',
			NOW(),
			JSON_OBJECT('IdMoneda', pIdMoneda, 'IdCuentaEmpresa', pIdCuentaEmpresa)
		);
	END IF;
    SELECT 'OK' Mensaje;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_crear_usuario` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_crear_usuario`(pUsuario varchar(30), pCredencial VARCHAR(255), pActor
   CHAR(10))
SALIR: BEGIN
      /*
      Permite crear un usuario administrativo en estado P: Pendiente.
      Genera una contraseña aleatoria que se devuelve para informar al usuario.
      Al iniciar sesión por primera vez, deberá cambiar su contraseña y se activará.
      Devuelve OK + Id + PasswordTemporal o el mensaje de error.
      Mensaje varchar(100), Id int, PasswordTemporal char(32)
      */
      DECLARE pIdUsuario INT;
      DECLARE pIdUsuarioActor INT DEFAULT NULL;
      DECLARE pTokenNuevo CHAR(32);
      DECLARE pPasswordTemporal char(32);
      -- Manejo de error en la transacción
      DECLARE EXIT HANDLER FOR SQLEXCEPTION
      BEGIN
          SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje, NULL Id, NULL PasswordTemporal;
      END;
      -- Resuelve identidad del actor
      IF pActor = 'USUARIO' THEN
          SET pIdUsuarioActor = f_valida_usuario(pCredencial);
          IF pIdUsuarioActor = 0 THEN
              SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje, NULL Id, NULL PasswordTemporal;
              LEAVE SALIR;
          END IF;
      END IF;
      -- Controla que el actor tenga rol Administrador
      IF pActor = 'USUARIO' AND (SELECT Rol FROM Usuarios WHERE IdUsuario = pIdUsuarioActor) != 'A' THEN
          SELECT 'No tienes permisos para realizar esta acción.' Mensaje, NULL Id, NULL PasswordTemporal;
          LEAVE SALIR;
      END IF;
      -- Controla parámetros
      IF (pUsuario IS NULL OR pUsuario = '') THEN
          SELECT 'El nombre de usuario es obligatorio.' Mensaje, NULL Id, NULL PasswordTemporal;
          LEAVE SALIR;
      END IF;
      IF EXISTS(SELECT Usuario FROM Usuarios WHERE Usuario = pUsuario) THEN
          SELECT 'El nombre de usuario ya existe.' Mensaje, NULL Id, NULL PasswordTemporal;
          LEAVE SALIR;
      END IF;
      -- Genera password y token de sesión de forma aleatoria
      SET pPasswordTemporal = CONCAT(
          SUBSTRING('ABCDEFGHJKLMNPQRSTUVWXYZ', FLOOR(1 + RAND() * 24), 1),
          SUBSTRING('abcdefghjkmnpqrstuvwxyz', FLOOR(1 + RAND() * 23), 1),
          SUBSTRING('0123456789', FLOOR(1 + RAND() * 10), 1),
          SUBSTRING('0123456789', FLOOR(1 + RAND() * 10), 1),
          SUBSTRING('ABCDEFGHJKLMNPQRSTUVWXYZ', FLOOR(1 + RAND() * 24), 1),
          SUBSTRING('abcdefghjkmnpqrstuvwxyz', FLOOR(1 + RAND() * 23), 1)
      );
      SET pTokenNuevo = md5(CONCAT(pUsuario, UNIX_TIMESTAMP(), RAND()));
      -- Crea al usuario
      INSERT INTO Usuarios (Usuario, `Password`, TokenSesion, FechaAlta, Estado)
      VALUES (pUsuario, md5(pPasswordTemporal), pTokenNuevo, NOW(), 'P');
      SET pIdUsuario = LAST_INSERT_ID();
      -- Audita
      INSERT INTO Operaciones (IdUsuario, TipoOperacion, FechaOperacion, Detalles)
      VALUES (pIdUsuarioActor, 'CU', NOW(), JSON_OBJECT('IdUsuario', pIdUsuario, 'Usuario', pUsuario));
      SELECT 'OK' Mensaje, pIdUsuario Id, pPasswordTemporal PasswordTemporal;
  END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_dame_moneda` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_dame_moneda`(
    pIdMoneda INT
)
SALIR: BEGIN
    /*
    Devuelve los datos de una moneda.
    */

    IF NOT EXISTS (SELECT 1 FROM Monedas WHERE IdMoneda = pIdMoneda) THEN
        SELECT 'La moneda no existe.' Mensaje,
               NULL IdMoneda, NULL IdCuentaEmpresa, NULL Estado, NULL FechaAlta;
        LEAVE SALIR;
    END IF;

    SELECT 'OK' Mensaje, IdMoneda, IdCuentaEmpresa, Estado, FechaAlta
    FROM Monedas
    WHERE IdMoneda = pIdMoneda;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_dame_parametro` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_dame_parametro`(pParametro varchar(50))
SALIR: BEGIN
	/*
    Devuelve los datos de un parámetro específico por su clave.
    */
    
     IF NOT EXISTS (SELECT 1 FROM Parametros WHERE Parametro = pParametro) THEN
        SELECT 'El parámetro no existe.' Mensaje, NULL Parametro, NULL Valor, NULL Descripcion, NULL EsModificable;
        LEAVE SALIR;
    END IF;
    
    SELECT	'OK' Mensaje, Parametro, Valor, Descripcion, EsModificable
    FROM	Parametros
    WHERE	Parametro = pParametro;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_dame_usuario` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_dame_usuario`(pIdUsuarioBuscado int)
SALIR: BEGIN
      /*
      Devuelve los datos de un usuario específico por su ID.
      */

      SELECT  'OK' Mensaje, IdUsuario, Usuario, FechaAlta, Estado, Rol
      FROM    Usuarios
      WHERE   IdUsuario = pIdUsuarioBuscado;
  END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_desactivar_moneda` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_desactivar_moneda`(
    pCredencial VARCHAR(255),
    pActor CHAR(10),
    pIdMoneda INT
)
SALIR: BEGIN
    /*
    Desactiva una moneda activa.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuario INT DEFAULT NULL;
    DECLARE pEstado CHAR(1);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
    END;

    IF pActor = 'USUARIO' THEN
        SET pIdUsuario = f_valida_usuario(pCredencial);
        IF pIdUsuario = 0 THEN
            SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
            LEAVE SALIR;
        END IF;
    END IF;

    SELECT Estado INTO pEstado
    FROM Monedas
    WHERE IdMoneda = pIdMoneda;

    IF pEstado IS NULL THEN
        SELECT 'La moneda no existe.' Mensaje;
        LEAVE SALIR;
    END IF;

    IF pEstado != 'A' THEN
        SELECT 'La moneda no está en estado Activo.' Mensaje;
        LEAVE SALIR;
    END IF;

    UPDATE Monedas
    SET Estado = 'I'
    WHERE IdMoneda = pIdMoneda;

    INSERT INTO Operaciones (IdUsuario, TipoOperacion, FechaOperacion, Detalles)
    VALUES (
        pIdUsuario,
        'DM',
        NOW(),
        JSON_OBJECT('IdMoneda', pIdMoneda)
    );

    SELECT 'OK' Mensaje;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_desactivar_usuario` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_desactivar_usuario`(pIdUsuario int, pCredencial VARCHAR(255), pActor
  CHAR(10))
SALIR: BEGIN
      /*
      Permite cambiar el estado de un usuario a I: Inactivo siempre y cuando esté activo.
      Devuelve OK o el mensaje de error en Mensaje.
      Mensaje varchar(100)
      */
      DECLARE pEstado char(1);
      DECLARE pIdUsuarioActor INT DEFAULT NULL;
      -- Manejo de error en la transacción
      DECLARE EXIT HANDLER FOR SQLEXCEPTION
      BEGIN
          SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
      END;
      -- Resuelve identidad del actor
      IF pActor = 'USUARIO' THEN
          SET pIdUsuarioActor = f_valida_usuario(pCredencial);
          IF pIdUsuarioActor = 0 THEN
              SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
              LEAVE SALIR;
          END IF;
      END IF;
      -- Controla que el actor tenga rol Administrador
      IF pActor = 'USUARIO' AND (SELECT Rol FROM Usuarios WHERE IdUsuario = pIdUsuarioActor) != 'A' THEN
          SELECT 'No tienes permisos para realizar esta acción.' Mensaje;
          LEAVE SALIR;
      END IF;
      -- Controla que el actor no se modifique a sí mismo
      IF pActor = 'USUARIO' AND pIdUsuarioActor = pIdUsuario THEN
          SELECT 'No puedes realizar esta acción sobre tu propia cuenta.' Mensaje;
          LEAVE SALIR;
      END IF;
      -- Controla parámetros
      SET pEstado = (SELECT Estado FROM Usuarios WHERE IdUsuario = pIdUsuario);
      IF pEstado IS NULL THEN
          SELECT 'El usuario no existe.' Mensaje;
          LEAVE SALIR;
      END IF;
      IF pEstado = 'I' THEN
          SELECT 'El usuario ya está inactivo.' Mensaje;
          LEAVE SALIR;
      END IF;
      -- Desactiva
      UPDATE  Usuarios
      SET     Estado = 'I',
              TokenSesion = md5(CONCAT(pIdUsuario, UNIX_TIMESTAMP()))
      WHERE   IdUsuario = pIdUsuario;
      -- Audita
      INSERT INTO Operaciones (IdUsuario, TipoOperacion, FechaOperacion, Detalles)
      VALUES (pIdUsuarioActor, 'DU', NOW(), JSON_OBJECT('IdUsuario', pIdUsuario));
      SELECT 'OK' Mensaje;
  END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_listar_monedas` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_listar_monedas`(pIncluyeInactivos char(1))
SALIR: BEGIN
    /*
    Permite listar monedas según pIncluyeInactivos:
      'N' → solo Activas (A).
      'S' → Activas e Inactivas (A, I).
      'T' → todas: Activas, Inactivas y Pendientes (A, I, P). Uso interno.
    Ordena por IdMoneda.
    */

    SET SESSION TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;

    SELECT      IdMoneda, IdCuentaEmpresa, Estado, FechaAlta
    FROM        Monedas
    WHERE       (pIncluyeInactivos = 'N' AND Estado = 'A')
             OR (pIncluyeInactivos = 'S' AND Estado IN ('A', 'I'))
             OR (pIncluyeInactivos = 'T' AND Estado IN ('A', 'I', 'P'))
    ORDER BY    IdMoneda;
     
    SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_login_usuario` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_login_usuario`(pUsuario varchar(30), pPassword char(32))
SALIR: BEGIN
      /*
      Permite a un usuario iniciar sesión en el sistema administrativo de MSTF.
      Valida credenciales, regenera el token de sesión y devuelve los datos del usuario.
      Si el usuario está Pendiente (Estado='P'), permite login pero debe cambiar contraseña.
      pPassword debe venir ya hasheado con md5 desde el cliente.
      Devuelve OK + datos del usuario o el mensaje de error en Mensaje.
      */
      DECLARE pIdUsuario INT;
      DECLARE pEstado CHAR(1);
      DECLARE pTokenSesion CHAR(32);

      DECLARE EXIT HANDLER FOR SQLEXCEPTION
      BEGIN
          SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje,
                 0 IdUsuario, NULL Usuario, NULL TokenSesion, NULL FechaAlta, NULL Estado, NULL Rol;
      END;

      -- Busca al usuario
      SELECT  IdUsuario, Estado
      INTO    pIdUsuario, pEstado
      FROM    Usuarios
      WHERE   Usuario = pUsuario AND `Password` = pPassword;

      -- Controla existencia
      IF pIdUsuario IS NULL THEN
          SELECT 'Credenciales inválidas.' Mensaje,
                 0 IdUsuario, NULL Usuario, NULL TokenSesion, NULL FechaAlta, NULL Estado, NULL Rol;
          LEAVE SALIR;
      END IF;

      -- Controla estado Inactivo
      IF pEstado = 'I' THEN
          SELECT 'El usuario está inactivo.' Mensaje,
                 0 IdUsuario, NULL Usuario, NULL TokenSesion, NULL FechaAlta, NULL Estado, NULL Rol;
          LEAVE SALIR;
      END IF;

      -- Regenera token de sesión
      SET pTokenSesion = md5(CONCAT(pIdUsuario, UNIX_TIMESTAMP(), RAND()));
      UPDATE  Usuarios
      SET     TokenSesion = pTokenSesion
      WHERE   IdUsuario = pIdUsuario;

      -- Devuelve datos del usuario (Estado='P' indica que debe cambiar contraseña)
      SELECT  'OK' Mensaje, IdUsuario, Usuario, pTokenSesion TokenSesion, FechaAlta, Estado, Rol
      FROM    Usuarios
      WHERE   IdUsuario = pIdUsuario;
  END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_logout_usuario` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_logout_usuario`(pTokenSesion char(32))
SALIR: BEGIN
	/*
    Permite a un usuario cerrar su sesión activa en el sistema administrativo de MSTF.
    Invalida el token actual generando uno nuevo desconocido para el cliente.
    Devuelve OK o el mensaje de error.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuario INT;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
    END;

    -- Busca el usuario activo por token
    SELECT  IdUsuario
    INTO    pIdUsuario
    FROM    Usuarios
    WHERE   TokenSesion = pTokenSesion AND Estado = 'A';

    IF pIdUsuario IS NULL THEN
        SELECT 'La sesión no existe o ya fue cerrada.' Mensaje;
        LEAVE SALIR;
    END IF;

    -- Invalida el token rotándolo a un valor nuevo desconocido para el cliente
    UPDATE  Usuarios
    SET     TokenSesion = md5(CONCAT(pIdUsuario, UNIX_TIMESTAMP(), RAND()))
    WHERE   IdUsuario = pIdUsuario;

    SELECT 'OK' Mensaje;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_modificar_parametro` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_modificar_parametro`(pCredencial VARCHAR(255), pActor CHAR(10), pParametro varchar(50), pValor text)
SALIR: BEGIN
	/*
    Permite modificar el valor de un parámetro siempre y cuando exista y sea modificable.
    Devuelve OK o el mensaje de error en Mensaje.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuario INT DEFAULT NULL;
    DECLARE pEsModificable, pEstado char(1);
    DECLARE pValorAnterior TEXT;
	-- Manejo de error en la transacción
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
		SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
	END;
    IF pActor = 'USUARIO' THEN
        SET pIdUsuario = f_valida_usuario(pCredencial);
        IF pIdUsuario = 0 THEN
            SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
            LEAVE SALIR;
        END IF;
    END IF;
    -- Controla parámetros
    SELECT	EsModificable,Valor
    INTO	pEsModificable, pValorAnterior
    FROM	Parametros 
    WHERE	Parametro = pParametro;
    
    IF pEsModificable IS NULL THEN
		SELECT 'El parámetro no existe.' Mensaje;
        LEAVE SALIR;
	END IF;
    
    IF pEsModificable = 'N' THEN
		SELECT 'El parámetro no es modificable desde el sitio administrativo.' Mensaje;
        LEAVE SALIR;
	END IF;
    IF (pValor IS NULL) THEN
        SELECT 'El valor es obligatorio.' Mensaje;
        LEAVE SALIR;
	END IF;
    
	-- Modifica el parámetro
	UPDATE	Parametros
    SET		Valor = pValor
    WHERE	Parametro = pParametro;
    
    -- Registra en auditoría
    INSERT INTO Operaciones (IdUsuario, TipoOperacion, FechaOperacion, Detalles)
    VALUES ( 
            pIdUsuario, 'MP', NOW(), 
            JSON_OBJECT('Parametro', pParametro, 'ValorAnterior', pValorAnterior, 'ValorNuevo', pValor));
	
	SELECT 'OK' Mensaje;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_modificar_password_usuario` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_modificar_password_usuario`(pTokenSesion char(32), pPasswordAnterior char(32), 
                                                   pPasswordNuevo varchar(50), pConfirmarPassword varchar(50))
SALIR: BEGIN
	/*
    Permite al usuario modificar su contraseña. Debe ingresar la contraseña anterior (hasheada con md5),
    la nueva y su confirmación.
    Devuelve OK o el mensaje de error en Mensaje.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuario INT;
	-- Manejo de error en la transacción
	DECLARE EXIT HANDLER FOR SQLEXCEPTION
	BEGIN
		SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
	END;
    SET pIdUsuario = f_valida_usuario(pTokenSesion);
    -- Verifica contraseña anterior
    IF NOT EXISTS(SELECT IdUsuario FROM Usuarios WHERE IdUsuario = pIdUsuario AND `Password` = pPasswordAnterior) THEN
        SELECT 'La contraseña anterior es incorrecta.' Mensaje;
        LEAVE SALIR;
	END IF;
    -- Controla parámetros de nueva contraseña
    IF (pPasswordNuevo IS NULL OR pPasswordNuevo = '') THEN
        SELECT 'La nueva contraseña es obligatoria.' Mensaje;
        LEAVE SALIR;
	END IF;
    IF (pConfirmarPassword IS NULL OR pConfirmarPassword = '') THEN
        SELECT 'La confirmación de contraseña es obligatoria.' Mensaje;
        LEAVE SALIR;
	END IF;
    IF (pPasswordNuevo != pConfirmarPassword) THEN
        SELECT 'La nueva contraseña no coincide con su confirmación.' Mensaje;
        LEAVE SALIR;
	END IF;
    
    -- Modifica la contraseña
    UPDATE	Usuarios
    SET		`Password` = pPasswordNuevo,
    TokenSesion = md5(CONCAT(pIdUsuario, UNIX_TIMESTAMP()))
    WHERE	IdUsuario = pIdUsuario;
    
    SELECT 'OK' Mensaje;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_restablecer_password_usuario` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_unicode_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_restablecer_password_usuario`(pIdUsuario int, pCredencial
  VARCHAR(255), pActor CHAR(10))
SALIR: BEGIN
      /*
      Permite restablecer la contraseña de un usuario. Solo puede hacerlo un Administrador.
      Genera una contraseña temporal, deja al usuario en estado Pendiente y regenera su token.
      Devuelve OK + PasswordTemporal o el mensaje de error.
      Mensaje varchar(100), PasswordTemporal varchar(10)
      */
      DECLARE pEstado CHAR(1);
      DECLARE pTokenNuevo CHAR(32);
      DECLARE pPasswordTemporal VARCHAR(10);
      DECLARE pIdUsuarioActor INT DEFAULT NULL;

      DECLARE EXIT HANDLER FOR SQLEXCEPTION
      BEGIN
          SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje, NULL PasswordTemporal;
      END;

      -- Resuelve identidad del actor
      IF pActor = 'USUARIO' THEN
          SET pIdUsuarioActor = f_valida_usuario(pCredencial);
          IF pIdUsuarioActor = 0 THEN
              SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje, NULL PasswordTemporal;
              LEAVE SALIR;
          END IF;
          -- Controla que el actor tenga rol Administrador
          IF (SELECT Rol FROM Usuarios WHERE IdUsuario = pIdUsuarioActor) != 'A' THEN
              SELECT 'No tienes permisos para realizar esta acción.' Mensaje, NULL PasswordTemporal;
              LEAVE SALIR;
          END IF;
      END IF;
      -- Controla parámetros del usuario objetivo
      SET pEstado = (SELECT Estado FROM Usuarios WHERE IdUsuario = pIdUsuario);
      IF pEstado IS NULL THEN
          SELECT 'El usuario no existe.' Mensaje, NULL PasswordTemporal;
          LEAVE SALIR;
      END IF;
      IF pEstado = 'P' THEN
          SELECT 'El usuario ya está en estado pendiente.' Mensaje, NULL PasswordTemporal;
          LEAVE SALIR;
      END IF;
      IF pEstado = 'I' THEN
          SELECT 'El usuario está inactivo.' Mensaje, NULL PasswordTemporal;
          LEAVE SALIR;
      END IF;

      -- Genera contraseña temporal (6 chars alfanuméricos)
      SET pPasswordTemporal = CONCAT(
          SUBSTRING('ABCDEFGHJKLMNPQRSTUVWXYZ', FLOOR(1 + RAND() * 24), 1),
          SUBSTRING('abcdefghjkmnpqrstuvwxyz', FLOOR(1 + RAND() * 23), 1),
          SUBSTRING('0123456789', FLOOR(1 + RAND() * 10), 1),
          SUBSTRING('0123456789', FLOOR(1 + RAND() * 10), 1),
          SUBSTRING('ABCDEFGHJKLMNPQRSTUVWXYZ', FLOOR(1 + RAND() * 24), 1),
          SUBSTRING('abcdefghjkmnpqrstuvwxyz', FLOOR(1 + RAND() * 23), 1)
      );

      -- Genera nuevo token
      SET pTokenNuevo = md5(CONCAT(pIdUsuario, UNIX_TIMESTAMP(), RAND()));

      -- Resetea Password, Token y Estado
      UPDATE  Usuarios
      SET     `Password` = md5(pPasswordTemporal),
              TokenSesion = pTokenNuevo,
              Estado = 'P'
      WHERE   IdUsuario = pIdUsuario;

      SELECT 'OK' Mensaje, pPasswordTemporal PasswordTemporal;
  END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

