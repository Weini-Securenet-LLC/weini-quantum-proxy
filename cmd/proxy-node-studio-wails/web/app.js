const state = {
  protocols: ["ss", "vmess", "vless", "trojan"],
  selected: new Set(["ss", "vmess", "vless", "trojan"]),
  nodes: [],
  lastResult: null,
  selectedURI: '',
  activeNode: null,
  proxyStatus: null,
};

const el = {
  urlInput: document.getElementById('urlInput'),
  fetchBtn: document.getElementById('fetchBtn'),
  openBtn: document.getElementById('openBtn'),
  connectBtn: document.getElementById('connectBtn'),
  disconnectBtn: document.getElementById('disconnectBtn'),
  protocolChips: document.getElementById('protocolChips'),
  protocolStats: document.getElementById('protocolStats'),
  nodesBody: document.getElementById('nodesBody'),
  trafficLogPanel: document.getElementById('trafficLogPanel'),
  totalNodes: document.getElementById('totalNodes'),
  progressBar: document.getElementById('progressBar'),
  statusText: document.getElementById('statusText'),
  activeNodeCard: document.getElementById('activeNodeCard'),
  proxyStatusCard: document.getElementById('proxyStatusCard'),
  toastHost: document.getElementById('toastHost'),
  titlebar: document.getElementById('titlebar'),
  windowMinBtn: document.getElementById('windowMinBtn'),
  windowCloseBtn: document.getElementById('windowCloseBtn'),
};

function showToast(message, type = 'success', duration = 2600) {
  const item = document.createElement('div');
  item.className = `toast ${type}`;
  item.textContent = message;
  el.toastHost.appendChild(item);
  setTimeout(() => {
    item.style.opacity = '0';
    item.style.transform = 'translateY(-8px)';
    setTimeout(() => item.remove(), 220);
  }, duration);
}

function appendTrafficLine(message) {
  const now = new Date().toLocaleTimeString();
  const line = `[${now}] ${message}`;
  const current = el.trafficLogPanel.textContent || '';
  const lines = `${current}\n${line}`.trim().split('\n');
  el.trafficLogPanel.textContent = lines.slice(-160).join('\n');
  el.trafficLogPanel.scrollTop = el.trafficLogPanel.scrollHeight;
}

function getErrorText(err) {
  if (!err) return '未知错误';
  if (typeof err === 'string') return err;
  if (typeof err.message === 'string' && err.message.trim()) return err.message;
  if (typeof err.error === 'string' && err.error.trim()) return err.error;
  try {
    return JSON.stringify(err);
  } catch {
    return String(err);
  }
}

function setProgress(value) {
  el.progressBar.style.width = `${value}%`;
}

function setStatus(message) {
  el.statusText.textContent = message;
}

function protocolPriority(protocol) {
  const idx = state.protocols.indexOf(protocol);
  return idx >= 0 ? idx : state.protocols.length;
}

function visibleNodes() {
  const indexed = state.nodes
    .map((node, index) => ({ node, index }))
    .filter(({ node }) => state.selected.has((node.protocol || '').toLowerCase()));
  indexed.sort((a, b) => {
    const diff = protocolPriority((a.node.protocol || '').toLowerCase()) - protocolPriority((b.node.protocol || '').toLowerCase());
    return diff !== 0 ? diff : a.index - b.index;
  });
  return indexed.map(item => item.node);
}

function currentSelectedNode() {
  const nodes = visibleNodes();
  return nodes.find(node => (node.raw_uri || '') === state.selectedURI) || nodes[0] || null;
}

function renderChips() {
  el.protocolChips.innerHTML = '';
  state.protocols.forEach(protocol => {
    const chip = document.createElement('button');
    chip.className = `chip ${state.selected.has(protocol) ? 'active' : ''}`;
    chip.textContent = protocol.toUpperCase();
    chip.onclick = () => {
      if (state.selected.has(protocol)) state.selected.delete(protocol);
      else state.selected.add(protocol);
      if (state.selected.size === 0) state.selected = new Set(state.protocols);
      const selectedNode = currentSelectedNode();
      state.selectedURI = selectedNode?.raw_uri || '';
      renderChips();
      renderNodes();
      if (selectedNode && state.activeNode?.raw_uri === selectedNode.raw_uri) {
        renderActiveNode(selectedNode);
      }
    };
    el.protocolChips.appendChild(chip);
  });
}

function renderStats(result = { protocol_counts: {} }) {
  el.protocolStats.innerHTML = '';
  state.protocols.forEach(protocol => {
    const li = document.createElement('li');
    li.innerHTML = `<span>${protocol.toUpperCase()}</span><strong>${result.protocol_counts?.[protocol] ?? 0}</strong>`;
    el.protocolStats.appendChild(li);
  });
}

function renderActiveNode(node = null) {
  state.activeNode = node;
  if (!node) {
    el.activeNodeCard.innerHTML = '未激活节点';
    return;
  }
  el.activeNodeCard.innerHTML = `
    <div><strong>${node.name || '未命名节点'}</strong></div>
    <div>${node.protocol.toUpperCase()} · ${node.host}:${node.port}</div>
    <div>传输: ${node.network || '-'}</div>
    <div>TLS: ${node.tls || '-'}</div>
    <div class="muted-line">退出程序时将自动断开代理并恢复系统代理。</div>
  `;
}

function renderProxyStatus(status = null) {
  state.proxyStatus = status;
  if (!status || !status.connected) {
    el.proxyStatusCard.innerHTML = `
      <div><strong>未连接</strong></div>
      <div class="muted-line">尚未连接全局代理</div>
      ${status?.check_message ? `<div class="muted-line">${status.check_message}</div>` : ''}
      ${status?.last_error ? `<div class="error-line">${status.last_error}</div>` : ''}
    `;
    return;
  }
  const node = status.node;
  const healthLabel = status.usable ? '可用' : '未验证';
  el.proxyStatusCard.innerHTML = `
    <div><strong>已连接 · ${status.mode?.toUpperCase() || 'TUN'}</strong></div>
    <div>状态: ${healthLabel}</div>
    <div>浏览器代理: ${status.system_proxy ? '已启用（127.0.0.1 本地 mixed）' : '未设置'}</div>
    <div>PID: ${status.pid || '-'}</div>
    <div>HTTP/混合端口: ${status.mixed_port || '-'}</div>
    <div>SOCKS 端口: ${status.socks_port || '-'}</div>
    ${status.check_message ? `<div class="muted-line">${status.check_message}</div>` : ''}
    ${node ? `<div class="muted-line">${node.protocol.toUpperCase()} · ${node.host}:${node.port}</div>` : ''}
  `;
}

function renderNodes() {
  const nodes = visibleNodes();
  el.nodesBody.innerHTML = '';
  nodes.forEach((node) => {
    const tr = document.createElement('tr');
    if ((node.raw_uri || '') === state.selectedURI) tr.classList.add('active');
    tr.innerHTML = `
      <td>${node.protocol.toUpperCase()}</td>
      <td>${node.name || '-'}</td>
      <td>${node.host}</td>
      <td>${node.port}</td>
      <td>${node.network || '-'}</td>
      <td>${node.tls || '-'}</td>
    `;
    tr.title = `${node.host}:${node.port}`;
    tr.onclick = () => {
      state.selectedURI = node.raw_uri || '';
      renderNodes();
    };
    tr.ondblclick = () => activateSelectedNode();
    el.nodesBody.appendChild(tr);
  });
}

function getWailsApp() {
  return window?.go?.wailsapp?.App || window?.go?.main?.App || null;
}

function getWailsRuntime() {
  return window?.runtime || window?.wails?.runtime || null;
}

function minimiseWindow() {
  const runtime = getWailsRuntime();
  if (runtime?.WindowMinimise) {
    runtime.WindowMinimise();
    return;
  }
  showToast('当前运行模式不支持最小化窗口控制', 'error');
}

function closeWindow() {
  const runtime = getWailsRuntime();
  if (runtime?.Quit) {
    runtime.Quit();
    return;
  }
  showToast('当前运行模式不支持窗口关闭控制', 'error');
}

async function fetchViaBridge(payload) {
  const bridge = getWailsApp();
  if (bridge?.FetchNodes) return bridge.FetchNodes(payload);
  const resp = await fetch('/api/fetch', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  const data = await resp.json();
  if (!resp.ok) throw new Error(data.error || '抓取失败');
  return data;
}

async function healthViaBridge() {
  const bridge = getWailsApp();
  if (bridge?.Health) return bridge.Health();
  const resp = await fetch('/api/health');
  return resp.json();
}

async function globalStatusViaBridge() {
  const bridge = getWailsApp();
  if (bridge?.GetGlobalProxyStatus) return bridge.GetGlobalProxyStatus();
  return null;
}

async function trafficLogViaBridge() {
  const bridge = getWailsApp();
  if (bridge?.GetTrafficLogSnapshot) return bridge.GetTrafficLogSnapshot();
  return { traffic_tail: '暂无流量日志', connected: false };
}

function renderTrafficLogs(snapshot = null) {
  el.trafficLogPanel.textContent = snapshot?.traffic_tail?.trim() || '暂无流量日志';
}

let statusPollHandle = null;

async function refreshRuntimePanels() {
  try {
    const [status, logs] = await Promise.all([globalStatusViaBridge(), trafficLogViaBridge()]);
    if (status) renderProxyStatus(status);
    renderTrafficLogs(logs);
  } catch (err) {
    appendTrafficLine(`错误：${getErrorText(err)}`);
  }
}

async function activateSelectedNode() {
  const selectedNode = currentSelectedNode();
  if (!selectedNode) {
    showToast('请先选中一个节点', 'error');
    return;
  }
  state.selectedURI = selectedNode.raw_uri || '';
  const uri = selectedNode.raw_uri || '';
  const bridge = getWailsApp();
  try {
    let node = selectedNode;
    if (bridge?.ActivateProxyURI) node = await bridge.ActivateProxyURI(uri);
    renderActiveNode(node);
    setStatus('节点已选择，可连接全局代理。');
    appendTrafficLine(`选择节点 ${node.host}:${node.port}`);
  } catch (err) {
    const message = getErrorText(err);
    appendTrafficLine(`错误：${message}`);
    showToast(message, 'error', 3200);
  }
}

async function connectGlobalProxy() {
  const bridge = getWailsApp();
  if (!bridge?.ConnectGlobalProxy) {
    showToast('当前运行模式不支持内置全局代理', 'error');
    return;
  }
  if (!state.activeNode) {
    await activateSelectedNode();
    if (!state.activeNode) return;
  }
  setStatus('正在连接全局代理...');
  appendTrafficLine('正在连接全局代理');
  try {
    const status = await bridge.ConnectGlobalProxy({ mode: 'tun', mixed_port: 7890, socks_port: 1080 });
    renderProxyStatus(status);
    await refreshRuntimePanels();
    setStatus(status.usable ? '全局代理已连接并验证通过' : '全局代理已启动');
    appendTrafficLine(`代理已连接 pid=${status.pid || 0}`);
  } catch (err) {
    const message = getErrorText(err);
    renderProxyStatus({ connected: false, last_error: message });
    setStatus('连接失败');
    appendTrafficLine(`错误：${message}`);
    showToast(message, 'error', 3600);
  }
}

async function disconnectGlobalProxy() {
  const bridge = getWailsApp();
  if (!bridge?.DisconnectGlobalProxy) {
    renderProxyStatus({ connected: false });
    return;
  }
  setStatus('正在断开代理...');
  appendTrafficLine('正在断开全局代理');
  try {
    const status = await bridge.DisconnectGlobalProxy();
    renderProxyStatus(status);
    setStatus('已成功断开链接');
    showToast('已成功断开链接', 'success');
    appendTrafficLine('已成功断开全局代理');
    refreshRuntimePanels();
  } catch (err) {
    const message = getErrorText(err);
    appendTrafficLine(`错误：${message}`);
    showToast(message, 'error', 3200);
  }
}

async function fetchNodes() {
  const payload = {
    url: el.urlInput.value.trim(),
    protocols: [...state.selected],
    timeout: 20,
  };
  setStatus('正在抓取...');
  setProgress(18);
  appendTrafficLine(`抓取 ${payload.url || '（默认地址）'}`);
  try {
    const data = await fetchViaBridge(payload);
    state.lastResult = data;
    state.nodes = data.nodes || [];
    state.selectedURI = state.nodes[0]?.raw_uri || '';
    el.totalNodes.textContent = String(data.total_nodes || 0);
    renderStats(data);
    renderNodes();
    renderActiveNode(null);
    setProgress(100);
    setStatus(`已加载 ${data.total_nodes} 个节点`);
    appendTrafficLine(`成功加载 ${data.total_nodes} 个节点`);
  } catch (err) {
    setStatus('抓取失败');
    setProgress(0);
    const message = getErrorText(err);
    appendTrafficLine(`错误：${message}`);
    showToast(message, 'error', 3200);
    return;
  }
  setTimeout(() => setProgress(0), 600);
}

async function bootstrap() {
  renderChips();
  renderStats();
  renderActiveNode(null);
  renderProxyStatus(null);
  renderTrafficLogs(null);
  if (el.titlebar) {
    el.titlebar.ondblclick = (event) => {
      event.preventDefault();
      event.stopPropagation();
      return false;
    };
  }
  if (el.windowMinBtn) el.windowMinBtn.onclick = minimiseWindow;
  if (el.windowCloseBtn) el.windowCloseBtn.onclick = closeWindow;
  el.fetchBtn.onclick = fetchNodes;
  el.openBtn.onclick = activateSelectedNode;
  el.connectBtn.onclick = connectGlobalProxy;
  el.disconnectBtn.onclick = disconnectGlobalProxy;
  try {
    const [health, status, logs] = await Promise.all([healthViaBridge(), globalStatusViaBridge(), trafficLogViaBridge()]);
    if (health?.default_url) el.urlInput.value = health.default_url;
    if (status) renderProxyStatus(status);
    renderTrafficLogs(logs);
    if (!statusPollHandle) statusPollHandle = setInterval(refreshRuntimePanels, 3000);
    appendTrafficLine('界面已就绪');
  } catch {
    appendTrafficLine('健康检查接口不可用');
  }
}

bootstrap();
