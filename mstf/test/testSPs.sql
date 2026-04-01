-- ----------------------------- --
-- Casos de prueba SPs MSTF      --
-- ----------------------------- --

-- 01. Autenticación
call tsp_autenticar_actor('APIKEY_INVALIDA', 'SISTEMA');
call tsp_autenticar_actor('CAMBIAR_ESTE_VALOR', 'SISTEMA');-- OK

call tsp_autenticar_actor('tokeninvalidoxxxxxxxxxxxxxxxxxxx', 'USUARIO');
call tsp_autenticar_actor((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- OK

-- 02. Usuarios
call tsp_buscar_usuarios('', 'S');
call tsp_buscar_usuarios('', 'N');
call tsp_buscar_usuarios('adm', 'S');
call tsp_buscar_usuarios('noexiste', 'S');

-- Crear usuario
call tsp_crear_usuario('', (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');
call tsp_crear_usuario('admin', (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');
call tsp_crear_usuario('usuario2', (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- OK
call tsp_crear_usuario('usuario2', (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- duplicado

-- Crear usuario via SISTEMA
call tsp_crear_usuario('usuario3', 'CAMBIAR_ESTE_VALOR', 'SISTEMA');-- OK
call tsp_crear_usuario('usuario3', 'CAMBIAR_ESTE_VALOR', 'SISTEMA');-- duplicado

-- Obtener usuario
call tsp_dame_usuario(1);
call tsp_dame_usuario(2);
call tsp_dame_usuario(3);
call tsp_dame_usuario(999);

-- Confirmar cuenta usuario2 (Estado P → A)
-- Nota: el servidor aplica md5() antes de llamar al SP. Las validaciones de longitud/letra/número
-- solo son testeables en la capa HTTP, ya que md5 siempre produce 32 hex chars válidos.
call tsp_confirmar_cuenta_usuario('tokeninvalidoxxxxxxxxxxxxxxxxxx', md5('user123'), md5('user123'));
call tsp_confirmar_cuenta_usuario((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 2), md5('user123'), md5('user456'));-- mismatch
call tsp_confirmar_cuenta_usuario((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 2), md5('user123'), md5('user123'));-- OK
call tsp_confirmar_cuenta_usuario((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 2), md5('user456'), md5('user456'));-- ya activo (idempotencia)

-- Confirmar cuenta usuario3
call tsp_confirmar_cuenta_usuario((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 3), md5('user123'), md5('user123'));-- OK

-- Login
call tsp_login_usuario('noexiste', md5('user123'));
call tsp_login_usuario('usuario2', md5('wrongpass'));
call tsp_login_usuario('usuario2', md5('user123'));-- OK (token se regenera)

-- Modificar password usuario2
call tsp_modificar_password_usuario('tokeninvalidoxxxxxxxxxxxxxxxxxx', md5('user123'), md5('newpass1'), md5('newpass1'));
call tsp_modificar_password_usuario((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 2), md5('wrongpass'), md5('newpass1'), md5('newpass1'));
call tsp_modificar_password_usuario((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 2), md5('user123'), md5('newpass1'), md5('newpass2'));-- mismatch
call tsp_modificar_password_usuario((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 2), md5('user123'), md5('newpass1'), md5('newpass1'));-- OK
call tsp_modificar_password_usuario((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 2), md5('newpass1'), md5('user123'), md5('user123'));-- OK (restaura)

-- Logout
call tsp_logout_usuario('tokeninvalidoxxxxxxxxxxxxxxxxxx');
set @tokenLogout = (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 2);
call tsp_logout_usuario(@tokenLogout);-- OK
call tsp_logout_usuario(@tokenLogout);-- sesión ya cerrada

-- Restablecer password (solo Admin)
call tsp_restablecer_password_usuario(999, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- no existe
call tsp_restablecer_password_usuario(2, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- OK (usuario2 vuelve a P)
call tsp_restablecer_password_usuario(2, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- ya pendiente
call tsp_confirmar_cuenta_usuario((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 2), md5('user123'), md5('user123'));-- OK (reactiva)

-- Desactivar usuario
call tsp_desactivar_usuario(1, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- automodificación
call tsp_desactivar_usuario(999, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- no existe
call tsp_desactivar_usuario(2, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- OK
call tsp_desactivar_usuario(2, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- ya inactivo

-- Activar usuario
call tsp_activar_usuario(1, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- automodificación
call tsp_activar_usuario(999, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- no existe
call tsp_activar_usuario(2, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- OK
call tsp_activar_usuario(2, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- ya activo

call tsp_buscar_usuarios('', 'S');

-- Borrar usuario
call tsp_borrar_usuario(1, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- automodificación
call tsp_borrar_usuario(999, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- no existe
call tsp_borrar_usuario(2, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- activo, no se puede
call tsp_desactivar_usuario(2, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- OK
call tsp_borrar_usuario(2, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- OK
call tsp_desactivar_usuario(3, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- OK
call tsp_borrar_usuario(3, (SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO');-- OK

call tsp_buscar_usuarios('', 'S');

-- 03. Parámetros
call tsp_buscar_parametros('', 'N');
call tsp_buscar_parametros('', 'S');
call tsp_buscar_parametros('LIMITE', 'N');

call tsp_dame_parametro('LIMITEBUSCARCUENTAS');
call tsp_dame_parametro('KAFKABATCHSIZE');
call tsp_dame_parametro('noexiste');

-- Modificar parámetro no modificable (EsModificable='N')
call tsp_modificar_parametro((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 'KAFKABATCHSIZE', '100');
-- Modificar parámetro inexistente
call tsp_modificar_parametro((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 'noexiste', '100');
-- Modificar sin valor
call tsp_modificar_parametro((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 'LIMITEBUSCARCUENTAS', NULL);
-- Modificar parámetro modificable
call tsp_modificar_parametro((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 'LIMITEBUSCARCUENTAS', '200');-- OK
call tsp_dame_parametro('LIMITEBUSCARCUENTAS');
call tsp_modificar_parametro((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 'LIMITEBUSCARCUENTAS', '500');-- OK restaura
-- Via SISTEMA
call tsp_modificar_parametro('CAMBIAR_ESTE_VALOR', 'SISTEMA', 'MONTOMAXTRANSFER', '200000');-- OK
call tsp_modificar_parametro('CAMBIAR_ESTE_VALOR', 'SISTEMA', 'MONTOMAXTRANSFER', '100000');-- OK restaura



-- 04. Monedas
call tsp_listar_monedas('N');-- solo activas
call tsp_listar_monedas('S');-- activas e inactivas
call tsp_listar_monedas('T');-- todas (A, I, P)

-- Crear moneda (Estado P: Pendiente)
call tsp_crear_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 0, 'cuenta-empresa-ars');-- id inválido
call tsp_crear_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', NULL, 'cuenta-empresa-ars');-- id nulo
call tsp_crear_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 1, '');-- sin IdCuentaEmpresa
call tsp_crear_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 1, NULL);-- IdCuentaEmpresa nulo
call tsp_crear_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 1, 'cuenta-empresa-ars');-- OK (P)
call tsp_crear_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 1, 'cuenta-empresa-ars');-- idempotente (ya en P, retorna OK)
call tsp_crear_moneda('CAMBIAR_ESTE_VALOR', 'SISTEMA', 2, 'cuenta-empresa-usd');-- OK
call tsp_crear_moneda('CAMBIAR_ESTE_VALOR', 'SISTEMA', 3, 'cuenta-empresa-eur');-- OK

call tsp_listar_monedas('T');-- 3 en estado P

-- Obtener moneda
call tsp_dame_moneda(1);
call tsp_dame_moneda(2);
call tsp_dame_moneda(999);

-- Activar moneda (P → A)
call tsp_activar_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 999);-- no existe
call tsp_activar_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 1);-- OK
call tsp_activar_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 1);-- ya activa
call tsp_activar_moneda('CAMBIAR_ESTE_VALOR', 'SISTEMA', 2);-- OK
call tsp_activar_moneda('CAMBIAR_ESTE_VALOR', 'SISTEMA', 3);-- OK

call tsp_listar_monedas('N');-- 3 activas

-- Desactivar moneda (A → I)
call tsp_desactivar_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 999);-- no existe
call tsp_desactivar_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 3);-- OK
call tsp_desactivar_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 3);-- ya inactiva

call tsp_listar_monedas('S');-- 1 y 2 activas, 3 inactiva

-- Borrar moneda (solo P o I)
call tsp_borrar_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 999);-- no existe
call tsp_borrar_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 1);-- activa, no se puede
call tsp_borrar_moneda((SELECT TokenSesion FROM Usuarios WHERE IdUsuario = 1), 'USUARIO', 3);-- OK (I → borrada)

call tsp_listar_monedas('S');-- quedan 1 y 2
call tsp_listar_monedas('T');-- igual, quedan 1 y 2
call tsp_listar_monedas('N');-- solo activas
call tsp_listar_monedas('S');-- activas e inactivas
call tsp_listar_monedas('T');-- todas (A, I, P)