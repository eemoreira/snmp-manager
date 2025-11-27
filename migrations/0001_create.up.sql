CREATE TABLE maquina (
  id INT AUTO_INCREMENT PRIMARY KEY,
  ip VARCHAR(45) NOT NULL,
  mac VARCHAR(50) NOT NULL,
  UNIQUE(ip),
  UNIQUE(mac)
) ENGINE=InnoDB;

CREATE TABLE switch_ (
  id INT AUTO_INCREMENT PRIMARY KEY,
  ip VARCHAR(45) NOT NULL,
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
  ip_maquina VARCHAR(45) NOT NULL,
  UNIQUE (sala_id, ip_maquina),
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
  acao ENUM('block', 'unblock') NOT NULL,
  tempo DATETIME NOT NULL,
  executado BOOLEAN NOT NULL DEFAULT FALSE,
  FOREIGN KEY (sala_id) REFERENCES sala(id)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB;

