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
-- Table structure for table `monedas`
--

DROP TABLE IF EXISTS `monedas`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `monedas` (
  `IdMoneda` int NOT NULL COMMENT 'PK de la tabla Monedas. Referencia el IdMoneda del sistema cliente.',
  `IdCuentaEmpresa` varchar(50) DEFAULT NULL,
  `FechaAlta` datetime NOT NULL COMMENT 'Fecha en que se creó la Moneda.',
  `Estado` char(1) NOT NULL COMMENT 'Estado de la Moneda: A (Activo) - I (Inactivo) - P (Pendiente)',
  PRIMARY KEY (`IdMoneda`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `operaciones`
--

DROP TABLE IF EXISTS `operaciones`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `operaciones` (
  `IdOperacion` int NOT NULL AUTO_INCREMENT COMMENT 'PK de la tabla Operaciones.',
  `IdUsuario` int NOT NULL COMMENT 'PK de la tabla Usuarios.',
  `TipoOperacion` char(2) NOT NULL COMMENT 'Tipo de operación que se audita: CC (creación de cuenta) - CT (creación de transferencia) - AC (activación de cuenta) - DA (desactivación de cuenta) - RT (reversión de transferencia)',
  `FechaOperacion` datetime NOT NULL,
  `Detalles` json NOT NULL,
  PRIMARY KEY (`IdOperacion`),
  KEY `Ref22` (`IdUsuario`),
  CONSTRAINT `RefUsuarios2` FOREIGN KEY (`IdUsuario`) REFERENCES `usuarios` (`IdUsuario`)
) ENGINE=InnoDB AUTO_INCREMENT=117 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Tabla que audita cualquier movimiento manual realizado por un Usuario sobre Cuentas o Transferencias en el MSTF.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `parametros`
--

DROP TABLE IF EXISTS `parametros`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `parametros` (
  `Parametro` varchar(50) NOT NULL COMMENT 'Nombre del Parametro.. Es único y PK.',
  `Valor` text NOT NULL COMMENT 'Valor del Parámetro.',
  `Descripcion` varchar(255) DEFAULT NULL COMMENT 'Descripción del parámetro. Opcional.',
  `EsModificable` char(1) NOT NULL COMMENT 'Define si un parámetro es del sistema o si puede ser modificado por un usuario administrativo. S (Si es modificable) - N (No es modificable)',
  PRIMARY KEY (`Parametro`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Tabla que almacena los Parámetros que se puedan definir con su valor.';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `usuarios`
--

DROP TABLE IF EXISTS `usuarios`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `usuarios` (
  `IdUsuario` int NOT NULL AUTO_INCREMENT COMMENT 'PK de la tabla Usuarios.',
  `Usuario` varchar(30) NOT NULL COMMENT 'Nombre de usuario del mismo. Es único.',
  `Password` char(32) NOT NULL COMMENT 'Clave encriptada en MD5.',
  `TokenSesion` char(32) NOT NULL COMMENT 'Token de sesión del cliente. Generado aleatoriamente y hasheado en MD5.',
  `FechaAlta` datetime NOT NULL COMMENT 'Fecha de creación del Usuario.',
  `Estado` char(1) NOT NULL COMMENT 'Estado del Usuario: A (Activo) - I (Inactivo)',
  PRIMARY KEY (`IdUsuario`),
  UNIQUE KEY `UI_Usuario` (`Usuario`),
  UNIQUE KEY `UI_TokenSesion` (`TokenSesion`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Tabla que almacena los Usuarios Administradores que gestionan aspectos del MSTF mediante el sitio administrativo.';
/*!40101 SET character_set_client = @saved_cs_client */;

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
 pTokenSesion CHAR(32),
    pIdMoneda INT
)
SALIR: BEGIN
    /*
    Activa una moneda en Estado Inactivo o Pendiente.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuario INT;
    DECLARE pEstado CHAR(1);
        DECLARE pLog VARCHAR(100);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
    END;
	SET pIdUsuario = f_valida_usuario(pTokenSesion);
    IF pIdUsuario = 0 THEN
        SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
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
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_activar_usuario`(pTokenSesion char(32), pIdUsuario int)
SALIR: BEGIN
	/*
    Permite cambiar el estado de un usuario a A: Activo siempre y cuando esté dado de baja.
    Devuelve OK o el mensaje de error en Mensaje.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuarioAdmin INT;
    DECLARE pEstado char(1);
	-- Manejo de error en la transacción
	DECLARE EXIT HANDLER FOR SQLEXCEPTION
	BEGIN
		SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
	END;
	-- Verifica si el usuario inició sesión
    SET pIdUsuarioAdmin = f_valida_usuario(pTokenSesion);
	IF pIdUsuarioAdmin = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
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
	UPDATE	Usuarios
    SET		Estado = 'A'
    WHERE	IdUsuario = pIdUsuario;
	
    SELECT 'OK' Mensaje;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_alta_operacion` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_alta_operacion`(pTokenSesion char(32), pTipoOperacion char(2), pDetalles json)
SALIR: BEGIN
	/*
    Permite registrar una operación de auditoría. Es llamado internamente por otros procedures
    o desde la aplicación para registrar operaciones realizadas en TigerBeetle.
    Devuelve OK + Id o el mensaje de error en Mensaje.
    Mensaje varchar(100), Id int
    */
    DECLARE pIdUsuario, pIdOperacion INT;
	-- Manejo de error en la transacción
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
		SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje, NULL Id;
	END;
	-- Verifica si el usuario inició sesión
    SET pIdUsuario = f_valida_usuario(pTokenSesion);
	IF pIdUsuario = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje, NULL Id;
        LEAVE SALIR;
	END IF;
    -- Controla parámetros
    IF (pTipoOperacion IS NULL OR pTipoOperacion = '') THEN
        SELECT 'El tipo de operación es obligatorio.' Mensaje, NULL Id;
        LEAVE SALIR;
	END IF;
    IF (pDetalles IS NULL) THEN
        SELECT 'Los detalles son obligatorios.' Mensaje, NULL Id;
        LEAVE SALIR;
	END IF;
    
	-- Calcula el próximo id
    SET pIdOperacion = (SELECT COALESCE(MAX(IdOperacion), 0) + 1 FROM Operaciones);
    -- Inserta el registro
    INSERT INTO Operaciones (IdOperacion, IdUsuario, TipoOperacion, FechaOperacion, Detalles)
    VALUES (pIdOperacion, pIdUsuario, pTipoOperacion, NOW(), pDetalles);
	
	SELECT 'OK' Mensaje, pIdOperacion Id;
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
    pTokenSesion CHAR(32),
    pIdMoneda INT
)
SALIR: BEGIN
    /*
    Borra una moneda únicamente si está en estado Inactivo o Pendiente.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuario INT;
    DECLARE pEstado CHAR(1);
    DECLARE pLog VARCHAR(100);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
    END;

    SET pIdUsuario = f_valida_usuario(pTokenSesion);
    IF pIdUsuario = 0 THEN
        SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
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
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_borrar_usuario`(pTokenSesion char(32), pIdUsuario int)
SALIR: BEGIN
	/*
    Permite eliminar un usuario siempre y cuando no tenga registros en Operaciones.
    No puede eliminarse a sí mismo.
    Devuelve OK o el mensaje de error en Mensaje.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuarioAdmin INT;
	-- Manejo de error en la transacción
	DECLARE EXIT HANDLER FOR SQLEXCEPTION
	BEGIN
		SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
	END;
	-- Verifica si el usuario inició sesión
    SET pIdUsuarioAdmin = f_valida_usuario(pTokenSesion);
	IF pIdUsuarioAdmin = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
	END IF;
    -- No puede eliminarse a sí mismo
    IF pIdUsuario = pIdUsuarioAdmin THEN
		SELECT 'No puede eliminarse a sí mismo.' Mensaje;
        LEAVE SALIR;
	END IF;
    -- Controla existencia
    IF NOT EXISTS(SELECT IdUsuario FROM Usuarios WHERE IdUsuario = pIdUsuario) THEN
		SELECT 'El usuario no existe.' Mensaje;
        LEAVE SALIR;
	END IF;
    -- Controla que no tenga registros en auditoría
    IF EXISTS(SELECT IdOperacion FROM Operaciones WHERE IdUsuario = pIdUsuario) THEN
		SELECT 'No se puede eliminar el usuario porque tiene operaciones registradas en auditoría.' Mensaje;
        LEAVE SALIR;
	END IF;
    
    -- Elimina
	DELETE FROM Usuarios WHERE IdUsuario = pIdUsuario;
	
    SELECT 'OK' Mensaje;
END ;;
DELIMITER ;
/*!50003 SET sql_mode              = @saved_sql_mode */ ;
/*!50003 SET character_set_client  = @saved_cs_client */ ;
/*!50003 SET character_set_results = @saved_cs_results */ ;
/*!50003 SET collation_connection  = @saved_col_connection */ ;
/*!50003 DROP PROCEDURE IF EXISTS `tsp_buscar_operacion` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_buscar_operacion`(pTokenSesion char(32), pIdUsuarioFiltro int, pTipoOperacion char(2), 
                                         pFechaDesde datetime, pFechaHasta datetime, pOffset int, pLimit int)
SALIR: BEGIN
	/*
    Permite buscar operaciones de auditoría con filtros opcionales: usuario, tipo de operación y 
    rango de fechas. Soporta paginación con offset y limit. Ordena por fecha descendente.
    Pasar NULL o 0 en los filtros para ignorarlos.
    */
    DECLARE pIdUsuario INT;
	-- Verifica si el usuario inició sesión
    SET pIdUsuario = f_valida_usuario(pTokenSesion);
	IF pIdUsuario = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
	END IF;
    
    -- Valores por defecto para paginación
    IF pOffset IS NULL OR pOffset < 0 THEN
		SET pOffset = 0;
	END IF;
    IF pLimit IS NULL OR pLimit <= 0 THEN
		SET pLimit = 50;
	END IF;
    
    SET SESSION TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
    
    SELECT		ao.IdOperacion, ao.IdUsuario, u.Usuario, ao.TipoOperacion, 
				ao.FechaOperacion, ao.Detalles
    FROM		Operaciones ao
    INNER JOIN	Usuarios u ON ao.IdUsuario = u.IdUsuario
    WHERE		(pIdUsuarioFiltro IS NULL OR pIdUsuarioFiltro = 0 OR ao.IdUsuario = pIdUsuarioFiltro) AND
				(pTipoOperacion IS NULL OR pTipoOperacion = '' OR ao.TipoOperacion = pTipoOperacion) AND
                (pFechaDesde IS NULL OR ao.FechaOperacion >= pFechaDesde) AND
                (pFechaHasta IS NULL OR ao.FechaOperacion <= pFechaHasta)
    ORDER BY	ao.FechaOperacion DESC
    LIMIT		pOffset, pLimit;
    
	SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;
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
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_buscar_parametros`(pTokenSesion char(32), pCadena varchar(50), pSoloModificables char(1))
SALIR: BEGIN
	/*
    Permite buscar los parámetros del sistema según su nombre. Si pSoloModificables es 'S', muestra solo los 
    modificables desde el sitio administrativo. Ordena por nombre de parámetro.
    */
    DECLARE pIdUsuario INT;
	-- Verifica si el usuario inició sesión
    SET pIdUsuario = f_valida_usuario(pTokenSesion);
	IF pIdUsuario = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
	END IF;
    
    SET SESSION TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
    
    SELECT		Parametro, Valor, Descripcion, EsModificable
    FROM		Parametros
    WHERE		Estado = 'A' AND (pSoloModificables = 'N' OR (EsModificable = 'S' AND pSoloModificables = 'S'))
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
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_buscar_usuarios`(pTokenSesion char(32), pCadena varchar(50), pIncluyeBajas char(1))
SALIR: BEGIN
	/*
    Permite listar todos los usuarios que cumplan con la condición de búsqueda: la cadena debe estar 
    contenida en el nombre de usuario. Puede o no incluir los usuarios dados de baja 
    según pIncluyeBajas (S: Si - N: No). Ordena por nombre de usuario.
    */
    DECLARE pIdUsuario INT;
	-- Verifica si el usuario inició sesión
    SET pIdUsuario = f_valida_usuario(pTokenSesion);
	IF pIdUsuario = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
	END IF;
    
    SET pCadena = COALESCE(pCadena, '');
    
    SET SESSION TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
    
    SELECT		IdUsuario, Usuario,FechaAlta, Estado
    FROM		Usuarios
    WHERE		(Usuario LIKE CONCAT('%', TRIM(pCadena), '%')) AND
				(pIncluyeBajas = 'S' OR (Estado != 'B' AND pIncluyeBajas = 'N'))
    ORDER BY	Usuario;
    
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
/*!50003 SET character_set_client  = utf8mb3 */ ;
/*!50003 SET character_set_results = utf8mb3 */ ;
/*!50003 SET collation_connection  = utf8mb3_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'NO_AUTO_VALUE_ON_ZERO' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_confirmar_cuenta_usuario`(pTokenSesion char(32), pPassword varchar(50), pConfirmarPassword varchar(50))
SALIR: BEGIN
	/*
    Permite al usuario Pendiente cambiar su contraseña temporal y activarse.
    Requiere haber iniciado sesión (tener token válido de tsp_login_usuario).
    Devuelve OK o el mensaje de error.
    Mensaje varchar(100)
    */
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
    -- Validaciones de contraseña
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
    SET     `Password` = md5(pPassword), 
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
    pTokenSesion CHAR(32),
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
    DECLARE pIdUsuario INT;
    DECLARE pEstado CHAR(1);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
    END;

    SET pIdUsuario = f_valida_usuario(pTokenSesion);
    IF pIdUsuario = 0 THEN
        SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
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
/*!50003 SET character_set_client  = utf8mb3 */ ;
/*!50003 SET character_set_results = utf8mb3 */ ;
/*!50003 SET collation_connection  = utf8mb3_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'NO_AUTO_VALUE_ON_ZERO' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_crear_usuario`(pTokenSesion char(32), pUsuario varchar(30))
SALIR: BEGIN
	/*
	Permite dar de alta un usuario administrativo en estado P: Pendiente.
    Genera una contraseña aleatoria que se devuelve para informar al usuario.
    Al iniciar sesión por primera vez, deberá cambiar su contraseña y se activará.
    Devuelve OK + Id + PasswordTemporal o el mensaje de error.
    Mensaje varchar(100), Id int, PasswordTemporal char(32)
    */
    DECLARE pIdUsuario INT;
    DECLARE pTokenNuevo CHAR(32);
    DECLARE pPasswordTemporal char(32);
	-- Manejo de error en la transacción
    DECLARE pIdUsuarioAdmin INT;
	-- Manejo de error en la transacción
	DECLARE EXIT HANDLER FOR SQLEXCEPTION
	BEGIN
		SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje, NULL Id, NULL PasswordTemporal;
	END;
	-- Verifica si el usuario inició sesión
    SET pIdUsuarioAdmin = f_valida_usuario(pTokenSesion);
	IF pIdUsuarioAdmin = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje, NULL Id, NULL PasswordTemporal;
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
    INSERT INTO Usuarios (Usuario, `Password`, TokenSesion,FechaAlta, Estado)
    VALUES (pUsuario, md5(pPasswordTemporal), pTokenNuevo, NOW(), 'P');
	SET pIdUsuario = LAST_INSERT_ID();
    
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
    pTokenSesion CHAR(32),
    pIdMoneda INT
)
SALIR: BEGIN
    /*
    Devuelve los datos de una moneda.
    */
    DECLARE pIdUsuario INT;

    SET pIdUsuario = f_valida_usuario(pTokenSesion);
    IF pIdUsuario = 0 THEN
        SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje,
               NULL IdMoneda,NULL IdCuentaEmpresa, NULL Estado, NULL FechaAlta;
        LEAVE SALIR;
    END IF;

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
/*!50003 DROP PROCEDURE IF EXISTS `tsp_dame_operacion` */;
/*!50003 SET @saved_cs_client      = @@character_set_client */ ;
/*!50003 SET @saved_cs_results     = @@character_set_results */ ;
/*!50003 SET @saved_col_connection = @@collation_connection */ ;
/*!50003 SET character_set_client  = utf8mb4 */ ;
/*!50003 SET character_set_results = utf8mb4 */ ;
/*!50003 SET collation_connection  = utf8mb4_0900_ai_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_dame_operacion`(pTokenSesion char(32), pIdOperacion int)
SALIR: BEGIN
	/*
    Devuelve los datos de una operación de auditoría específica por su ID.
    */
    DECLARE pIdUsuario INT;
	-- Verifica si el usuario inició sesión
    SET pIdUsuario = f_valida_usuario(pTokenSesion);
	IF pIdUsuario = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
	END IF;
    
    SELECT		ao.IdOperacion, ao.IdUsuario, u.Usuario, ao.TipoOperacion, 
				ao.FechaOperacion, ao.Detalles
    FROM		Operaciones ao
    INNER JOIN	Usuarios u ON ao.IdUsuario = u.IdUsuario
    WHERE		ao.IdOperacion = pIdOperacion;
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
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_dame_parametro`(pTokenSesion char(32), pParametro varchar(50))
SALIR: BEGIN
	/*
    Devuelve los datos de un parámetro específico por su clave.
    */
    DECLARE pIdUsuario INT;
	-- Verifica si el usuario inició sesión
    SET pIdUsuario = f_valida_usuario(pTokenSesion);
	IF pIdUsuario = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje, NULL Parametro, NULL Valor, NULL Descripcion, NULL EsModificable;
        LEAVE SALIR;
	END IF;
    
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
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_dame_usuario`(pTokenSesion char(32), pIdUsuarioBuscado int)
SALIR: BEGIN
	/*
    Devuelve los datos de un usuario específico por su ID.
    */
    DECLARE pIdUsuario INT;
	-- Verifica si el usuario inició sesión
    SET pIdUsuario = f_valida_usuario(pTokenSesion);
	IF pIdUsuario = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje, NULL, NULL, NULL, NULL;
        LEAVE SALIR;
	END IF;
    
    SELECT	'OK' Mensaje, IdUsuario, Usuario, FechaAlta, Estado
    FROM	Usuarios
    WHERE	IdUsuario = pIdUsuarioBuscado;
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
    pTokenSesion CHAR(32),
    pIdMoneda INT
)
SALIR: BEGIN
    /*
    Desactiva una moneda activa.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuario INT;
    DECLARE pEstado CHAR(1);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
    END;

    SET pIdUsuario = f_valida_usuario(pTokenSesion);
    IF pIdUsuario = 0 THEN
        SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
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
/*!50003 SET character_set_client  = utf8mb3 */ ;
/*!50003 SET character_set_results = utf8mb3 */ ;
/*!50003 SET collation_connection  = utf8mb3_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'NO_AUTO_VALUE_ON_ZERO' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_desactivar_usuario`(pTokenSesion char(32), pIdUsuario int)
SALIR: BEGIN
	/*
    Permite cambiar el estado de un usuario a I: Inactivo siempre y cuando no esté desactivado.
    No puede desactivarse a sí mismo.
    Devuelve OK o el mensaje de error en Mensaje.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuarioAdmin INT;
    DECLARE pEstado char(1);
	-- Manejo de error en la transacción
	DECLARE EXIT HANDLER FOR SQLEXCEPTION
	BEGIN
		SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
	END;
	-- Verifica si el usuario inició sesión
    SET pIdUsuarioAdmin = f_valida_usuario(pTokenSesion);
	IF pIdUsuarioAdmin = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
	END IF;
    -- No puede darse de baja a sí mismo
    IF pIdUsuarioAdmin = pIdUsuario THEN
		SELECT 'No puede darse de baja a sí mismo.' Mensaje;
        LEAVE SALIR;
	END IF;
    -- Controla parámetros
    SET pEstado = (SELECT Estado FROM Usuarios WHERE IdUsuario = pIdUsuario);
    IF pEstado IS NULL THEN
		SELECT 'El usuario no existe.' Mensaje;
        LEAVE SALIR;
	END IF;
    IF pEstado = 'B' THEN
		SELECT 'El usuario ya está dado de baja.' Mensaje;
        LEAVE SALIR;
	END IF;
    
    -- Da de baja
	UPDATE	Usuarios
    SET		Estado = 'B',
            TokenSesion = md5(CONCAT(pIdUsuario, UNIX_TIMESTAMP()))
    WHERE	IdUsuario = pIdUsuario;
	
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
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_listar_monedas`(pTokenSesion char(32), pIncluyeInactivas char(1))
SALIR: BEGIN
    /*
    Permite listar todas las monedas. Si pIncluyeInactivas es 'S', muestra todas.
    Si es 'N', muestra solo las activas. Ordena por IdMoneda.
    */
    DECLARE pIdUsuario INT;
    
    SET pIdUsuario = f_valida_usuario(pTokenSesion);
    IF pIdUsuario = 0 THEN
        SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
    END IF;
    
    SET SESSION TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
    
    SELECT      IdMoneda,  IdCuentaEmpresa, Estado, FechaAlta
    FROM        Monedas
    WHERE       (pIncluyeInactivas = 'S' OR Estado != 'I') AND Estado IN('A', 'I', 'P')
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
    Si el usuario está Pendiente, permite login pero indica que debe cambiar contraseña.
    pPassword debe venir ya hasheado con md5 desde el cliente.
    Devuelve OK + datos del usuario o el mensaje de error en Mensaje.
    */
    DECLARE pIdUsuario INT;
    DECLARE pEstado CHAR(1);
    DECLARE pTokenSesion CHAR(32);
    DECLARE pRequiereCambio CHAR(1) DEFAULT 'N';
    
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje, 
               0 IdUsuario, NULL Usuario, NULL TokenSesion, NULL RequiereCambioPassword, NULL FechaAlta, NULL Estado;
    END;
    
    -- Busca al usuario
    SELECT  IdUsuario, Estado
    INTO    pIdUsuario, pEstado
    FROM    Usuarios 
    WHERE   Usuario = pUsuario AND `Password` = pPassword;
    
    -- Controla existencia
    IF pIdUsuario IS NULL THEN
        SELECT 'Credenciales inválidas.' Mensaje, 
               0 IdUsuario, NULL Usuario, NULL TokenSesion, NULL RequiereCambioPassword, NULL FechaAlta, NULL Estado;
        LEAVE SALIR;
    END IF;
    
    -- Controla estado Baja
    IF pEstado = 'B' THEN
        SELECT 'El usuario está dado de baja.' Mensaje, 
               0 IdUsuario, NULL Usuario, NULL TokenSesion, NULL RequiereCambioPassword, NULL FechaAlta, NULL Estado;
        LEAVE SALIR;
    END IF;
    
    -- Si está Pendiente, permite login pero indica que debe cambiar contraseña
    IF pEstado = 'P' THEN
        SET pRequiereCambio = 'S';
    END IF;
    
    -- Regenera token de sesión
    SET pTokenSesion = md5(CONCAT(pIdUsuario, UNIX_TIMESTAMP(), RAND()));
    UPDATE  Usuarios
    SET     TokenSesion = pTokenSesion
    WHERE   IdUsuario = pIdUsuario;
    
    -- Devuelve datos del usuario
    SELECT  'OK' Mensaje, IdUsuario, Usuario, pTokenSesion TokenSesion, pRequiereCambio RequiereCambioPassword, FechaAlta, Estado
    FROM    Usuarios
    WHERE   IdUsuario = pIdUsuario;
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
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_modificar_parametro`(pTokenSesion char(32), pParametro varchar(50), pValor text)
SALIR: BEGIN
	/*
    Permite modificar el valor de un parámetro siempre y cuando exista y sea modificable.
    Devuelve OK o el mensaje de error en Mensaje.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuario INT;
    DECLARE pEsModificable, pEstado char(1);
    DECLARE pValorAnterior TEXT;
	-- Manejo de error en la transacción
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
		SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
	END;
	-- Verifica si el usuario inició sesión
    SET pIdUsuario = f_valida_usuario(pTokenSesion);
	IF pIdUsuario = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
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
    la nueva y su confirmación. La política de contraseñas establece que ésta debe tener una longitud 
    mínima de 6 caracteres y debe incluir por lo menos una letra y un número. 
    Devuelve OK o el mensaje de error en Mensaje.
    Mensaje varchar(100)
    */
    DECLARE pIdUsuario INT;
	-- Manejo de error en la transacción
	DECLARE EXIT HANDLER FOR SQLEXCEPTION
	BEGIN
		SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje;
	END;
	-- Verifica si el usuario inició sesión
    SET pIdUsuario = f_valida_usuario(pTokenSesion);
	IF pIdUsuario = 0 THEN
		SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje;
        LEAVE SALIR;
	END IF;
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
    IF (CHAR_LENGTH(pPasswordNuevo) < 6) THEN
        SELECT 'La contraseña debe tener al menos 6 caracteres.' Mensaje;
        LEAVE SALIR;
	END IF;
    IF (pPasswordNuevo NOT REGEXP '[A-Za-z]') THEN
        SELECT 'La contraseña debe incluir al menos una letra.' Mensaje;
        LEAVE SALIR;
	END IF;
    IF (pPasswordNuevo NOT REGEXP '[0-9]') THEN
        SELECT 'La contraseña debe incluir al menos un número.' Mensaje;
        LEAVE SALIR;
	END IF;
    IF (pPasswordNuevo != pConfirmarPassword) THEN
        SELECT 'La nueva contraseña no coincide con su confirmación.' Mensaje;
        LEAVE SALIR;
	END IF;
    
    -- Modifica la contraseña
    UPDATE	Usuarios
    SET		`Password` = md5(pPasswordNuevo),
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
/*!50003 SET character_set_client  = utf8mb3 */ ;
/*!50003 SET character_set_results = utf8mb3 */ ;
/*!50003 SET collation_connection  = utf8mb3_general_ci */ ;
/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'NO_AUTO_VALUE_ON_ZERO' */ ;
DELIMITER ;;
CREATE DEFINER=`root`@`localhost` PROCEDURE `tsp_restablecer_password_usuario`(pTokenSesion char(32), pIdUsuario int)
SALIR: BEGIN
	/*
    Permite a un administrador logueado restablecer la contraseña de otro usuario. 
    Genera una contraseña temporal, deja al usuario en estado Pendiente y regenera su token.
    Devuelve OK + PasswordTemporal o el mensaje de error.
    Mensaje varchar(100), PasswordTemporal varchar(10)
    */
    DECLARE pIdUsuarioAdmin INT;
    DECLARE pEstado CHAR(1);
    DECLARE pTokenNuevo CHAR(32);
    DECLARE pPasswordTemporal VARCHAR(10);
    
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        SELECT 'Error en la transacción. Contáctese con el administrador.' Mensaje, NULL PasswordTemporal;
    END;
    
    -- Verifica si el usuario inició sesión
    SET pIdUsuarioAdmin = f_valida_usuario(pTokenSesion);
    IF pIdUsuarioAdmin = 0 THEN
        SELECT 'La sesión expiró. Vuelva a iniciar sesión.' Mensaje, NULL PasswordTemporal;
        LEAVE SALIR;
    END IF;
    
    -- No puede restablecerse a sí mismo por este método
    IF pIdUsuarioAdmin = pIdUsuario THEN
        SELECT 'No puede restablecer su propia contraseña por este método.' Mensaje, NULL PasswordTemporal;
        LEAVE SALIR;
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
    IF pEstado = 'B' THEN
        SELECT 'El usuario está dado de baja.' Mensaje, NULL PasswordTemporal;
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

-- Dump completed on 2026-02-24 22:55:39
