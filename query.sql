-- 1. Cria o Switch
INSERT INTO switch_ (ip) VALUES ('10.90.90.90');

-- Pega o ID do Switch criado
SET @switch_id = LAST_INSERT_ID(); 

-- 2. Cria a Porta Admin no Switch
-- Assumimos que o IP 10.90.90.51 está conectado na Porta 7 do Switch.
INSERT INTO porta_switch (switch_id, porta_numero) VALUES (@switch_id, 7);

-- Pega o ID da Porta criada
SET @porta_admin_id = LAST_INSERT_ID();

-- 3. Cria a Máquina Admin
-- O MAC pode ser qualquer valor válido, neste exemplo: 00:1A:2B:3C:4D:5E
INSERT INTO maquina (ip, mac) VALUES ('10.90.90.60', '00:1A:2B:3C:4D:5E');

-- Pega o ID da Máquina criada
SET @maquina_admin_id = LAST_INSERT_ID();

-- 4. Cria a Sala de Teste (que será controlada pela Máquina Admin)
-- Use 'admin' / 'senha' para testar o login na interface.
INSERT INTO sala (nome, maquina_admin_id, login_admin, senha_admin) 
VALUES ('Sala Teste', @maquina_admin_id, 'admin', 'senha');

-- Pega o ID da Sala criada
SET @sala_id = LAST_INSERT_ID();

-- 5. Associa a Sala à Porta e ao IP da Máquina Admin
-- Esta é a relação que o backend usa para saber se o IP que acessa (10.90.90.51)
-- tem permissão para gerenciar a sala associada a essa porta.
INSERT INTO sala_porta (sala_id, porta_switch_id, ip_maquina) VALUES (@sala_id, @porta_admin_id, '10.90.90.60');

-- Opcional: Adicionar uma segunda porta para testar as funções de LIGAR/DESLIGAR
INSERT INTO porta_switch (switch_id, porta_numero) VALUES (@switch_id, 9);
SET @porta_other_id = LAST_INSERT_ID();

INSERT INTO sala_porta (sala_id, porta_switch_id, ip_maquina) VALUES (@switch_id, @porta_other_id, "");

-- Verifica os dados inseridos (opcional)
SELECT * FROM maquina;
SELECT * FROM switch_;
SELECT * FROM porta_switch;
SELECT * FROM sala;
SELECT * FROM sala_porta;
