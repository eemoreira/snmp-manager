const API_BASE = '/api';
let machines = [];
let refreshIntervalId = null;
let currentTab = 'control';
let lastControlAction = 'down';
let lastBatchAction = 'down';

window.onload = async () => {
    try {
        const response = await fetch(`${API_BASE}/check`);
        if (response.status === 200) {
            document.getElementById('loading')?.classList.add('hidden');
            document.getElementById('login')?.classList.remove('hidden');
        } else {
            document.body.innerHTML = "<div style='text-align: center; padding: 50px;'><h1>Acesso Negado</h1><p>IP não autorizado</p></div>";
        }
    } catch (e) {
        console.error('Erro ao verificar acesso:', e);
        document.body.innerHTML = "<div style='text-align: center; padding: 50px;'><h1>Erro de Conexão</h1><p>Não foi possível conectar ao servidor</p></div>";
    }
};

async function doLogin() {
    const login = document.getElementById('loginUser')?.value;
    const password = document.getElementById('loginPass')?.value;
    const msg = document.getElementById('loginMsg');

    try {
        const res = await fetch(`${API_BASE}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ login, password })
        });

        if (res.ok) {
            document.getElementById('login')?.classList.add('hidden');
            document.getElementById('dashboard')?.classList.remove('hidden');
            const welcomeEl = document.getElementById('welcomeMsg');
            if (welcomeEl) welcomeEl.textContent = `Bem-vindo, ${login}`;
            
            await loadMachines();
            startAutoRefresh();
            
            showMessage(document.getElementById('dashMsg'), 'Login realizado com sucesso', 'success');
        } else {
            showMessage(msg, 'Credenciais inválidas', 'error');
        }
    } catch (e) {
        console.error('Erro no login:', e);
        showMessage(msg, 'Erro de conexão', 'error');
    }
}

const validateFutureDate = (datetime) => Math.floor((new Date(datetime) - new Date()) / 1000);

const toggleDisplay = (el, show) => el.style.display = show ? 'block' : 'none';

const updateButton = (btn, text, className) => {
    btn.textContent = text;
    btn.className = className;
};

function updateActionFields() {
    const action = document.querySelector('input[name="ctrlAction"]:checked')?.value;
    const isBlocking = action === 'down';
    const autoUnblockInput = document.getElementById('autoUnblockDateTime');
    
    lastControlAction = action;
    
    toggleDisplay(document.getElementById('blockFields'), isBlocking);
    toggleDisplay(document.getElementById('unblockFields'), !isBlocking);
    autoUnblockInput.required = isBlocking;
    
    updateButton(document.getElementById('executeBtn'), 
        isBlocking ? 'Bloquear com Desbloqueio Automático' : 'Desbloquear Agora',
        isBlocking ? 'danger' : 'success');
    
    if (!isBlocking) updateScheduleUnblockFields();
}

function updateScheduleUnblockFields() {
    const isScheduled = document.getElementById('scheduleUnblock')?.checked;
    const scheduleInput = document.getElementById('scheduleUnblockDateTime');
    
    toggleDisplay(document.getElementById('scheduleUnblockFields'), isScheduled);
    scheduleInput.required = isScheduled;
    
    updateButton(document.getElementById('executeBtn'),
        isScheduled ? 'Agendar Desbloqueio' : 'Desbloquear Agora',
        isScheduled ? 'warning' : 'success');
}

function updateBatchFields() {
    const action = document.querySelector('input[name="batchAction"]:checked')?.value;
    const isBlocking = action === 'down';
    const autoUnblockInput = document.getElementById('batchAutoUnblockDateTime');
    
    lastBatchAction = action;
    
    toggleDisplay(document.getElementById('batchBlockFields'), isBlocking);
    toggleDisplay(document.getElementById('batchUnblockFields'), !isBlocking);
    autoUnblockInput.required = isBlocking;
    
    updateButton(document.getElementById('batchBtn'),
        isBlocking ? 'Bloquear Todas' : 'Desbloquear Todas',
        isBlocking ? 'danger' : 'success');
    
    if (!isBlocking) updateBatchScheduleFields();
}

function updateBatchScheduleFields() {
    const isScheduled = document.getElementById('scheduleBatchUnblock')?.checked;
    const scheduleInput = document.getElementById('batchScheduleUnblockDateTime');
    
    toggleDisplay(document.getElementById('batchScheduleUnblockFields'), isScheduled);
    scheduleInput.required = isScheduled;
    
    updateButton(document.getElementById('batchBtn'),
        isScheduled ? 'Agendar Desbloqueio de Todas' : 'Desbloquear Todas',
        isScheduled ? 'warning' : 'success');
}

async function executeControl(event) {
    event.preventDefault();
    const ip = document.getElementById('ctrlIP')?.value;
    const action = document.querySelector('input[name="ctrlAction"]:checked')?.value;
    const msg = document.getElementById('dashMsg');
    
    lastControlAction = action;

    if (action === 'down') {
        const diffInSeconds = validateFutureDate(document.getElementById('autoUnblockDateTime')?.value);
        if (diffInSeconds <= 0) {
            showMessage(msg, 'A data/hora de desbloqueio deve ser no futuro', 'error');
            return;
        }
        const unblockTime = new Date(document.getElementById('autoUnblockDateTime').value);

        try {
            // 1. Bloqueia imediatamente
            const blockRes = await fetch(`${API_BASE}/ports`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ip, state: 'down' })
            });

            if (!blockRes.ok) {
                const err = await blockRes.json();
                showMessage(msg, 'Erro ao bloquear: ' + (err.error || 'Desconhecido'), 'error');
                return;
            }

            // 2. Agenda desbloqueio automático
            const scheduleRes = await fetch(`${API_BASE}/agendamento`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ 
                    ip, 
                    state: 'up',
                    time_offset: diffInSeconds.toString()
                })
            });

            if (scheduleRes.ok) {
                showMessage(msg, `Porta bloqueada. Desbloqueio automático em ${unblockTime.toLocaleString('pt-BR')}`, 'success');
                document.getElementById('individualForm')?.reset();
                document.querySelector(`input[name="ctrlAction"][value="${lastControlAction}"]`).checked = true;
                updateActionFields();
                await loadMachines();
                await loadSchedules();
            } else {
                showMessage(msg, 'Porta bloqueada, mas falha ao agendar desbloqueio. IMPORTANTE: Desbloquear manualmente!', 'error');
            }
        } catch (e) {
            console.error('Erro ao bloquear:', e);
            showMessage(msg, 'Erro de rede', 'error');
        }

    } else {
        const isScheduled = document.getElementById('scheduleUnblock')?.checked;

        if (isScheduled) {
            const datetime = document.getElementById('scheduleUnblockDateTime')?.value;
            const diffInSeconds = validateFutureDate(datetime);
            if (diffInSeconds <= 0) {
                showMessage(msg, 'A data/hora deve ser no futuro', 'error');
                return;
            }
            const scheduleTime = new Date(datetime);

            try {
                const res = await fetch(`${API_BASE}/agendamento`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ 
                        ip, 
                        state: 'up',
                        time_offset: diffInSeconds.toString()
                    })
                });

                if (res.ok) {
                    showMessage(msg, `Desbloqueio agendado para ${scheduleTime.toLocaleString('pt-BR')}`, 'success');
                    document.getElementById('individualForm')?.reset();
                    document.querySelector(`input[name="ctrlAction"][value="${lastControlAction}"]`).checked = true;
                    const checkbox = document.getElementById('scheduleUnblock');
                    if (checkbox) checkbox.checked = false;
                    updateScheduleUnblockFields();
                    await loadMachines();
                    await loadSchedules();
                } else {
                    const err = await res.json();
                    showMessage(msg, 'Erro: ' + (err.error || 'Desconhecido'), 'error');
                }
            } catch (e) {
                console.error('Erro ao agendar:', e);
                showMessage(msg, 'Erro de rede', 'error');
            }
        } else {
            try {
                const res = await fetch(`${API_BASE}/ports`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ ip, state: 'up' })
                });

                if (res.ok) {
                    showMessage(msg, 'Porta desbloqueada com sucesso', 'success');
                    document.getElementById('individualForm')?.reset();
                    document.querySelector(`input[name="ctrlAction"][value="${lastControlAction}"]`).checked = true;
                    updateActionFields();
                    await loadMachines();
                    await loadSchedules();
                } else {
                    const err = await res.json();
                    showMessage(msg, 'Erro: ' + (err.error || 'Desconhecido'), 'error');
                }
            } catch (e) {
                console.error('Erro ao desbloquear:', e);
                showMessage(msg, 'Erro de rede', 'error');
            }
        }
    }
}

async function executeBatchControl() {
    const action = document.querySelector('input[name="batchAction"]:checked')?.value;
    const msg = document.getElementById('dashMsg');

    if (action === 'down') {
        const autoUnblockDateTime = document.getElementById('batchAutoUnblockDateTime')?.value;
        const diffInSeconds = validateFutureDate(autoUnblockDateTime);
        
        if (diffInSeconds <= 0) {
            showMessage(msg, 'A data/hora de desbloqueio deve ser no futuro', 'error');
            return;
        }
        const unblockTime = new Date(autoUnblockDateTime);

        if (!confirm(`Deseja bloquear TODAS as máquinas?\nDesbloqueio automático em: ${unblockTime.toLocaleString('pt-BR')}`)) return;

        showMessage(msg, 'Bloqueando todas as máquinas...', 'info');

        let successCount = 0;
        let errorCount = 0;
        let scheduleErrors = 0;

        for (const machine of machines) {
            try {
                const blockRes = await fetch(`${API_BASE}/ports`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ ip: machine.ip, state: 'down' })
                });

                if (blockRes.ok) {
                    const scheduleRes = await fetch(`${API_BASE}/agendamento`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ 
                            ip: machine.ip, 
                            state: 'up',
                            time_offset: diffInSeconds.toString()
                        })
                    });

                    if (scheduleRes.ok) {
                        successCount++;
                    } else {
                        successCount++;
                        scheduleErrors++;
                    }
                } else {
                    errorCount++;
                }
            } catch (e) {
                errorCount++;
            }
        }

        let message = `Bloqueadas: ${successCount}, Erros: ${errorCount}`;
        if (scheduleErrors > 0) {
            message += ` (${scheduleErrors} sem agendamento - DESBLOQUEAR MANUALMENTE!)`;
        }
        
        showMessage(msg, message, errorCount > 0 || scheduleErrors > 0 ? 'error' : 'success');
        document.getElementById('batchAutoUnblockDateTime').value = '';
        await loadMachines();
    } else {
        const isScheduled = document.getElementById('scheduleBatchUnblock')?.checked;

        if (isScheduled) {
            const datetime = document.getElementById('batchScheduleUnblockDateTime')?.value;
            const diffInSeconds = validateFutureDate(datetime);
            
            if (diffInSeconds <= 0) {
                showMessage(msg, 'A data/hora deve ser no futuro', 'error');
                return;
            }
            const scheduleTime = new Date(datetime);

            if (!confirm(`Agendar desbloqueio de TODAS as máquinas para ${scheduleTime.toLocaleString('pt-BR')}?`)) return;

            showMessage(msg, 'Agendando desbloqueio...', 'info');

            let successCount = 0;
            let errorCount = 0;

            for (const machine of machines) {
                try {
                    const res = await fetch(`${API_BASE}/agendamento`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ 
                            ip: machine.ip, 
                            state: 'up',
                            time_offset: diffInSeconds.toString()
                        })
                    });

                    if (res.ok) {
                        successCount++;
                    } else {
                        errorCount++;
                    }
                } catch (e) {
                    errorCount++;
                }
            }

            showMessage(msg, `Agendamentos criados: ${successCount}, Erros: ${errorCount}`, 
                        errorCount > 0 ? 'error' : 'success');
            document.getElementById('batchScheduleUnblockDateTime').value = '';
            const checkbox = document.getElementById('scheduleBatchUnblock');
            if (checkbox) checkbox.checked = false;
            updateBatchScheduleFields();
        } else {
            if (!confirm('Deseja desbloquear TODAS as máquinas agora?')) return;

            showMessage(msg, 'Desbloqueando todas as máquinas...', 'info');

            let successCount = 0;
            let errorCount = 0;

            for (const machine of machines) {
                try {
                    const res = await fetch(`${API_BASE}/ports`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ ip: machine.ip, state: 'up' })
                    });

                    if (res.ok) {
                        successCount++;
                    } else {
                        errorCount++;
                    }
                } catch (e) {
                    errorCount++;
                }
            }

            showMessage(msg, `Desbloqueadas: ${successCount}, Erros: ${errorCount}`, 
                        errorCount > 0 ? 'error' : 'success');
            
            await loadMachines();
        }
    }
}

async function createMachine(event) {
    event.preventDefault();
    const ip = document.getElementById('newIP')?.value;
    const mac = document.getElementById('newMAC')?.value;
    const porta_num = parseInt(document.getElementById('newPort')?.value);
    const msg = document.getElementById('registerMsg');

    try {
        const res = await fetch(`${API_BASE}/maquinas`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ip, mac, porta_num })
        });

        if (res.status === 201) {
            showMessage(msg, 'Máquina cadastrada com sucesso', 'success');
            document.getElementById('registerForm')?.reset();
            await loadMachines();
        } else {
            const err = await res.json();
            showMessage(msg, 'Erro: ' + (err.error || 'Falha ao criar'), 'error');
        }
    } catch (e) {
        console.error('Erro ao criar máquina:', e);
        showMessage(msg, 'Erro de rede', 'error');
    }
}

async function loadMachines() {
    const machinesList = document.getElementById('machinesList');
    if (!machinesList) return;

    try {
        const res = await fetch(`${API_BASE}/maquinas`);
        if (res.ok) {
            const data = await res.json();
            machines = data.machines || [];
            renderMachines();
        } else {
            machinesList.innerHTML = '<p class="message error">Erro ao carregar máquinas</p>';
        }
    } catch (e) {
        console.error('Erro ao carregar máquinas:', e);
        machinesList.innerHTML = '<p class="message error">Erro de conexão</p>';
    }
}

function renderMachines() {
    const container = document.getElementById('machinesList');
    if (!container) return;

    if (machines.length === 0) {
        container.innerHTML = '<p class="message info">Nenhuma máquina cadastrada</p>';
        return;
    }

    container.innerHTML = machines.map(machine => {
        const isActive = machine.status === 1;
        const statusClass = isActive ? 'active' : 'blocked';
        const statusText = isActive ? 'Ativa' : 'Bloqueada';

        return `
            <div class="machine-card ${statusClass}">
                <div class="machine-info">
                    <div class="machine-name">Máquina ${machine.ip}</div>
                    <div class="machine-details">
                        IP: ${machine.ip} | MAC: ${machine.mac} | Porta: ${machine.porta_num}
                    </div>
                </div>
                <div style="display: flex; align-items: center; gap: 15px;">
                    <span class="status-badge ${statusClass}">${statusText}</span>
                    <label class="toggle-switch" title="Clique para ${isActive ? 'bloquear' : 'desbloquear'}">
                        <input type="checkbox" ${isActive ? 'checked' : ''} 
                               onchange="toggleMachineSwitch('${machine.ip}', this.checked)">
                        <span class="slider"></span>
                    </label>
                </div>
            </div>
        `;
    }).join('');
}

async function toggleMachineSwitch(ip, isChecked) {
    // Na aba de máquinas, só permite DESBLOQUEAR (ativar)
    if (!isChecked) {
        // Tentou bloquear - não permitido aqui
        alert('Use a aba "Controle" para bloquear máquinas.\n\nA aba "Máquinas" permite apenas desbloquear.');
        await loadMachines(); // Restaura o estado visual
        return;
    }
    
    // Desbloquear (up) - execução imediata
    try {
        const res = await fetch(`${API_BASE}/ports`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ip, state: 'up' })
        });

        if (res.ok) {
            await loadMachines();
            await loadSchedules();
        } else {
            await loadMachines();
        }
    } catch (e) {
        console.error('Erro ao desbloquear:', e);
        await loadMachines();
    }
}

function startAutoRefresh() {
    if (refreshIntervalId) clearInterval(refreshIntervalId);

    refreshIntervalId = setInterval(async () => {
        if (currentTab === 'machines') await loadMachines();
        else if (currentTab === 'schedules') await loadSchedules();
    }, 10000);
}

function updateControlMode() {
    const isIndividual = document.querySelector('input[name="controlMode"]:checked')?.value === 'individual';
    toggleDisplay(document.getElementById('individualControl'), isIndividual);
    toggleDisplay(document.getElementById('batchControl'), !isIndividual);
}

function switchTab(tabName) {
    currentTab = tabName;

    document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
    event.target.classList.add('active');

    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    document.getElementById(`${tabName}Tab`)?.classList.add('active');

    if (tabName === 'machines') loadMachines();
    else if (tabName === 'schedules') loadSchedules();
}

async function loadSchedules() {
    const schedulesList = document.getElementById('schedulesList');
    if (!schedulesList) return;

    schedulesList.innerHTML = '<div class="spinner"></div>';

    try {
        const res = await fetch(`${API_BASE}/agendamentos`);
        if (res.ok) {
            const data = await res.json();
            renderSchedules(data.schedules || []);
        } else {
            schedulesList.innerHTML = '<p class="message error">Erro ao carregar agendamentos</p>';
        }
    } catch (e) {
        console.error('Erro ao carregar agendamentos:', e);
        schedulesList.innerHTML = '<p class="message error">Erro de conexão</p>';
    }
}

function renderSchedules(schedules) {
    const container = document.getElementById('schedulesList');
    if (!container) return;

    if (schedules.length === 0) {
        container.innerHTML = '<p class="message info">Nenhum agendamento pendente</p>';
        return;
    }

    container.innerHTML = schedules.map(schedule => {
        const execTime = new Date(schedule.ExecutarEm);
        const actionText = schedule.Acao === 'up' ? 'Desbloquear' : 'Bloquear';
        const actionClass = schedule.Acao === 'up' ? 'success' : 'danger';

        return `
            <div class="machine-card">
                <div class="machine-info">
                    <div class="machine-name">${actionText} ${schedule.MaquinaIP}</div>
                    <div class="machine-details">
                        Programado para: ${execTime.toLocaleString('pt-BR')}
                    </div>
                </div>
                <span class="status-badge ${actionClass}">${actionText}</span>
            </div>
        `;
    }).join('');
}

function showMessage(element, text, type) {
    if (!element) return;
    
    element.textContent = text;
    element.className = `message ${type}`;
    element.style.display = 'block';
    
    if (type === 'success') {
        setTimeout(() => element.style.display = 'none', 5000);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('loginPass')?.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') doLogin();
    });
    
    document.getElementById('scheduleUnblock')?.addEventListener('change', updateScheduleUnblockFields);
});
