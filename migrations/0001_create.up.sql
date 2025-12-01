CREATE TABLE maquina (
  id INT AUTO_INCREMENT PRIMARY KEY,
  ip VARCHAR(45) NOT NULL,
  mac VARCHAR(50) NOT NULL,
  UNIQUE(ip),
  UNIQUE(mac)
) ENGINE=InnoDB;

CREATE TABLE switch_ (
  id INT AUTO_INCREMENT PRIMARY KEY,
  ip VARCHAR(45) NOT NULL
) ENGINE=InnoDB;

CREATE TABLE porta_switch (
  id INT AUTO_INCREMENT PRIMARY KEY,
  switch_id INT NOT NULL,
  porta_numero INT NOT NULL,
  FOREIGN KEY (switch_id) REFERENCES switch_(id)
    ON DELETE CASCADE ON UPDATE CASCADE,
  UNIQUE (switch_id, porta_numero)
) ENGINE=InnoDB;

CREATE TABLE sala (
  id INT AUTO_INCREMENT PRIMARY KEY,
  nome VARCHAR(100) NOT NULL,
  maquina_admin_id INT NOT NULL,
  login_admin VARCHAR(100) NOT NULL,
  senha_admin VARCHAR(255) NOT NULL,
  FOREIGN KEY (maquina_admin_id) REFERENCES maquina(id)
    ON DELETE RESTRICT ON UPDATE CASCADE,
  UNIQUE (nome)
) ENGINE=InnoDB;

CREATE TABLE sala_porta (
  sala_id INT NOT NULL,
  porta_switch_id INT NOT NULL,
  ip_maquina VARCHAR(45),
  PRIMARY KEY (sala_id, porta_switch_id),
  FOREIGN KEY (sala_id) REFERENCES sala(id)
    ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (porta_switch_id) REFERENCES porta_switch(id)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB;

CREATE TABLE agendamento (
  id INT AUTO_INCREMENT PRIMARY KEY,
  sala_id INT NOT NULL,
  ip_maquina VARCHAR(45) NOT NULL,
  acao ENUM('up', 'down') NOT NULL,
  executar_em DATETIME NOT NULL,
  executado BOOLEAN NOT NULL DEFAULT FALSE,
  FOREIGN KEY (sala_id) REFERENCES sala(id)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB;

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

