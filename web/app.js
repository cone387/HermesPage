(function() {
    const $ = id => document.getElementById(id);
    const content = $('content');
    const emptyEl = $('empty');
    const searchInput = $('search');
    const categoryButtons = $('category-buttons');
    const tagButtons = $('tag-buttons');
    const ownerButtons = $('owner-buttons');
    const ownerFilterRow = $('owner-filter-row');
    const userInfoEl = $('user-info');
    const userMenu = $('user-menu');
    const userMenuTrigger = $('user-menu-trigger');
    const userDropdown = $('user-dropdown');
    const logoutBtn = $('logout-btn');
    const sortSelect = $('sort-select');
    const themeToggle = $('theme-toggle');
    const uploadBtn = $('upload-btn');
    const uploadModal = $('upload-modal');
    let allReports = [];
    let categories = [];
    let ownerMap = {};
    let selectedCategory = '';
    let selectedTag = '';
    let selectedOwner = '';
    let sortBy = 'time-desc';
    let loaded = false;
    let token = localStorage.getItem('token');
    let user;
    try { user = JSON.parse(localStorage.getItem('user') || 'null'); } catch (_) { user = null; }
    if (!token) user = null;

    const categoryLabel = value => value;
    const icons = {
        file: '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8Z M14 2v6h6 M8 13h8 M8 17h5"/>',
        lock: '<rect x="5" y="10" width="14" height="11" rx="2"/><path d="M8 10V7a4 4 0 0 1 8 0v3"/>',
        globe: '<circle cx="12" cy="12" r="9"/><path d="M3 12h18 M12 3a17 17 0 0 1 0 18 17 17 0 0 1 0-18Z"/>',
        trash: '<path d="M3 6h18 M9 6V3h6v3 M5 6l1 15h12l1-15 M10 10v7 M14 10v7"/>'
    };
    const icon = name => '<svg viewBox="0 0 24 24" aria-hidden="true">' + icons[name] + '</svg>';

    async function init() {
        try {
            const resp = await fetch('/api/setup/status');
            const status = await resp.json();
            if (status.needs_setup) { window.location.href = '/setup.html'; return; }
        } catch (_) {}
        // Check the session before showing controls for private data.
        if (token) {
            try {
                const resp = await fetch('/api/auth/me', {headers: authHeaders()});
                if (resp.ok) {
                    user = await resp.json();
                    localStorage.setItem('user', JSON.stringify({id: user.id, username: user.username, role: user.role}));
                } else if (resp.status === 401) {
                    token = null; user = null;
                    localStorage.removeItem('token'); localStorage.removeItem('user');
                }
            } catch (_) {}
        }
        if (user) {
            userMenu.style.display = 'block';
            userMenuTrigger.textContent = user.username;
            uploadBtn.style.display = 'inline-flex';
        } else {
            userInfoEl.innerHTML = '<a href="/login.html">登录</a>';
        }
        userMenuTrigger.addEventListener('click', () => {
            const open = userDropdown.classList.toggle('show');
            userMenuTrigger.setAttribute('aria-expanded', String(open));
        });
        document.addEventListener('click', e => {
            if (!userMenu.contains(e.target)) closeUserMenu();
        });
        document.addEventListener('keydown', e => {
            if (e.key === 'Escape') closeUserMenu();
            if (e.key === '/' && !e.ctrlKey && !e.metaKey && !e.altKey &&
                !e.target.closest('input, textarea, select, [contenteditable="true"]') &&
                !document.querySelector('dialog[open]')) {
                e.preventDefault(); searchInput.focus();
            }
        });
        logoutBtn.addEventListener('click', () => {
            localStorage.removeItem('token'); localStorage.removeItem('user'); window.location.reload();
        });
        themeToggle.addEventListener('click', () => {
            const next = document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark';
            document.documentElement.dataset.theme = next;
            localStorage.setItem('theme', next);
            updateThemeLabel();
        });
        updateThemeLabel();
        sortSelect.addEventListener('change', () => { sortBy = sortSelect.value; render(); });
        uploadBtn.addEventListener('click', () => uploadModal.showModal());
        $('clear-filters').addEventListener('click', clearFilters);
        $('empty-action').addEventListener('click', () => {
            if (hasFilters()) clearFilters();
            else if (user) uploadModal.showModal();
            else window.location.href = '/login.html';
        });
        initUpload();
        await fetchReports();
    }

    function closeUserMenu() {
        userDropdown.classList.remove('show');
        userMenuTrigger.setAttribute('aria-expanded', 'false');
    }

    function updateThemeLabel() {
        const label = document.documentElement.dataset.theme === 'dark' ? '切换为浅色主题' : '切换为深色主题';
        themeToggle.setAttribute('aria-label', label);
        themeToggle.title = label;
    }

    let noticeTimer;
    function notify(message) {
        const el = $('notice');
        clearTimeout(noticeTimer);
        el.textContent = message; el.hidden = false;
        noticeTimer = setTimeout(() => { el.hidden = true; }, 5000);
    }

    async function request(url, options = {}) {
        const resp = await fetch(url, options);
        if (!resp.ok) {
            const data = await resp.json().catch(() => ({}));
            throw new Error(data.error || (resp.status === 401 ? '登录已失效，请重新登录。' : '操作失败，请稍后重试。'));
        }
        return resp;
    }

    function initUpload() {
        const closeBtn = uploadModal.querySelector('.modal-close');
        const dropZone = $('drop-zone');
        const fileInput = $('file-input');
        const form = $('upload-form');
        const error = $('upload-error');
        const submit = form.querySelector('[type="submit"]');
        let uploading = false;
        closeBtn.addEventListener('click', () => { if (!uploading) uploadModal.close(); });
        uploadModal.addEventListener('cancel', e => { if (uploading) e.preventDefault(); });
        uploadModal.addEventListener('click', e => {
            if (e.target === uploadModal && !uploading) {
                const rect = uploadModal.getBoundingClientRect();
                if (e.clientX < rect.left || e.clientX > rect.right || e.clientY < rect.top || e.clientY > rect.bottom) uploadModal.close();
            }
        });
        const showFile = () => {
            error.hidden = true;
            dropZone.querySelector('.drop-text').textContent = fileInput.files[0]?.name || '拖拽 HTML 文件到此处，或点击选择';
        };
        dropZone.addEventListener('dragover', e => { e.preventDefault(); if (!uploading) dropZone.classList.add('dragover'); });
        dropZone.addEventListener('dragleave', () => dropZone.classList.remove('dragover'));
        dropZone.addEventListener('drop', e => {
            e.preventDefault(); dropZone.classList.remove('dragover');
            if (!uploading && e.dataTransfer.files.length) { fileInput.files = e.dataTransfer.files; showFile(); }
        });
        fileInput.addEventListener('change', showFile);
        form.addEventListener('submit', async e => {
            e.preventDefault();
            if (uploading || !fileInput.files.length) return;
            if (!/\.html?$/i.test(fileInput.files[0].name)) {
                error.textContent = '请选择 .html 或 .htm 格式的报告。'; error.hidden = false; return;
            }
            const data = new FormData();
            data.append('file', fileInput.files[0]);
            ['title', 'category', 'tags', 'visibility'].forEach(key => {
                const value = $('upload-' + key).value.trim();
                if (value) data.append(key, value);
            });
            uploading = true; error.hidden = true;
            submit.disabled = true; closeBtn.disabled = true; fileInput.disabled = true;
            submit.textContent = '正在上传…';
            try {
                await request('/api/upload', {method: 'POST', headers: authHeaders(), body: data});
                uploadModal.close(); form.reset(); showFile();
                await fetchReports(); notify('报告上传成功');
            } catch (err) { error.textContent = err.message; error.hidden = false; }
            finally {
                uploading = false; submit.disabled = false; closeBtn.disabled = false; fileInput.disabled = false;
                submit.textContent = '上传报告';
            }
        });
    }

    function authHeaders() { return token ? {Authorization: 'Bearer ' + token} : {}; }

    async function fetchReports() {
        content.setAttribute('aria-busy', 'true');
        try {
            const resp = await request('/api/list', {headers: authHeaders()});
            const data = await resp.json();
            allReports = data.reports || [];
            categories = data.categories || [];
            ownerMap = data.owners || {};
            loaded = true;
            renderFilters(); render();
        } catch (_) {
            loaded = false;
            emptyEl.style.display = 'none';
            $('result-count').textContent = '报告暂时无法加载';
            content.innerHTML = '<div class="empty"><h2>加载失败</h2><p>请检查网络连接，然后重试。</p><button id="retry-load" class="secondary-button">重新加载</button></div>';
            $('retry-load').addEventListener('click', fetchReports);
        } finally { content.setAttribute('aria-busy', 'false'); }
    }

    function getAllTags() {
        const tags = new Set();
        allReports.forEach(r => (r.tags || []).forEach(tag => tags.add(tag)));
        return [...tags].sort();
    }

    function isDateTag(tag) {
        if (!/^\d{4}-\d{2}-\d{2}$/.test(tag)) return false;
        const date = new Date(tag + 'T00:00:00Z');
        return !Number.isNaN(date.getTime()) && date.toISOString().slice(0, 10) === tag;
    }

    function partitionTags(tags) {
        const topics = [];
        const dates = [];
        tags.forEach(tag => (isDateTag(tag) ? dates : topics).push(tag));
        dates.sort((a, b) => b.localeCompare(a));
        return {topics, dates};
    }

    function tagColorClass(tag) {
        let hash = 0;
        for (let i = 0; i < tag.length; i++) hash = tag.charCodeAt(i) + ((hash << 5) - hash);
        return 'tag-' + (Math.abs(hash) % 6);
    }

    function renderFilters() {
        if (!categories.includes(selectedCategory)) selectedCategory = '';
        const tags = getAllTags();
        if (!tags.includes(selectedTag)) selectedTag = '';
        if (!ownerMap[selectedOwner]) selectedOwner = '';
        categoryButtons.innerHTML = [['', '全部'], ...categories.slice().sort().map(cat => [cat, cat])]
            .map(([value, label]) => '<button class="cat-btn" data-cat="' + escapeHtml(value) + '">' + escapeHtml(label) + '</button>').join('');
        const {topics, dates} = partitionTags(tags);
        tagButtons.innerHTML = [''].concat(topics).map(tag => '<button class="tag-btn" data-tag="' + escapeHtml(tag) + '">' + escapeHtml(tag || '全部') + '</button>').join('');
        if (dates.length) {
            tagButtons.insertAdjacentHTML('beforeend', '<select id="date-tag-select" class="date-tag-select" aria-label="日期标签"><option value="">日期标签 · ' + dates.length + '</option>' + dates.map(date => '<option value="' + date + '">' + date + '</option>').join('') + '</select>');
        }
        ownerFilterRow.style.display = user?.role === 'admin' && Object.keys(ownerMap).length ? 'flex' : 'none';
        ownerButtons.innerHTML = [['', '全部'], ...Object.entries(ownerMap)]
            .map(([id, name]) => '<button class="tag-btn" data-owner="' + escapeHtml(id) + '">' + escapeHtml(name) + '</button>').join('');
        syncFilters();
    }

    function syncFilters() {
        [[categoryButtons, 'cat', selectedCategory], [tagButtons, 'tag', selectedTag], [ownerButtons, 'owner', selectedOwner]].forEach(([container, key, selected]) => {
            container.querySelectorAll('button').forEach(btn => {
                const active = btn.dataset[key] === selected;
                btn.classList.toggle('active', active);
                btn.setAttribute('aria-pressed', String(active));
            });
        });
        $('clear-filters').hidden = !hasFilters();
        const dateSelect = $('date-tag-select');
        if (dateSelect) {
            const selectedDate = isDateTag(selectedTag) ? selectedTag : '';
            dateSelect.value = selectedDate;
            dateSelect.classList.toggle('active', Boolean(selectedDate));
        }
    }

    function hasFilters() { return Boolean(selectedCategory || selectedTag || selectedOwner || searchInput.value.trim()); }
    function clearFilters() {
        selectedCategory = ''; selectedTag = ''; selectedOwner = '';
        searchInput.value = '';
        render();
    }

    categoryButtons.addEventListener('click', e => {
        const btn = e.target.closest('button[data-cat]');
        if (btn) { selectedCategory = btn.dataset.cat; render(); }
    });
    tagButtons.addEventListener('click', e => {
        const btn = e.target.closest('button[data-tag]');
        if (btn) { selectedTag = btn.dataset.tag; render(); }
    });
    tagButtons.addEventListener('change', e => {
        if (e.target.id === 'date-tag-select') {
            selectedTag = e.target.value;
            render();
        }
    });
    ownerButtons.addEventListener('click', e => {
        const btn = e.target.closest('button[data-owner]');
        if (btn) { selectedOwner = btn.dataset.owner; render(); }
    });

    function getFiltered() {
        const search = searchInput.value.toLowerCase().trim();
        const filtered = allReports.filter(r => {
            if (selectedCategory && r.category !== selectedCategory) return false;
            if (selectedTag && !(r.tags || []).includes(selectedTag)) return false;
            if (selectedOwner && r.owner !== selectedOwner) return false;
            return !search || [r.title, r.summary, r.category, categoryLabel(r.category), ...(r.tags || [])]
                .some(value => String(value || '').toLowerCase().includes(search));
        });
        filtered.sort((a, b) => {
            switch (sortBy) {
                case 'time-asc': return new Date(a.created_at) - new Date(b.created_at);
                case 'title-asc': return a.title.localeCompare(b.title, 'zh-CN');
                case 'title-desc': return b.title.localeCompare(a.title, 'zh-CN');
                case 'size-desc': return b.size - a.size;
                case 'size-asc': return a.size - b.size;
                default: return new Date(b.created_at) - new Date(a.created_at);
            }
        });
        return filtered;
    }

    function groupByDate(reports) {
        if (!sortBy.startsWith('time-')) return {'全部报告': reports};
        const now = new Date();
        const today = dateKey(now);
        const yesterday = new Date(now); yesterday.setDate(now.getDate() - 1);
        const weekAgo = new Date(now); weekAgo.setDate(now.getDate() - 7);
        const labels = ['今天', '昨天', '近 7 天', '更早'];
        if (sortBy === 'time-asc') labels.reverse();
        const groups = Object.fromEntries(labels.map(label => [label, []]));
        reports.forEach(r => {
            const date = new Date(r.created_at);
            const key = dateKey(date);
            const label = key === today ? '今天' : key === dateKey(yesterday) ? '昨天' : date > weekAgo ? '近 7 天' : '更早';
            groups[label].push(r);
        });
        return groups;
    }

    function dateKey(date) {
        return date.getFullYear() + '-' + String(date.getMonth() + 1).padStart(2, '0') + '-' + String(date.getDate()).padStart(2, '0');
    }
    function formatTime(value) {
        const date = new Date(value);
        const now = new Date();
        if (dateKey(date) === dateKey(now)) return date.toLocaleTimeString('zh-CN', {hour: '2-digit', minute: '2-digit'});
        return date.toLocaleDateString('zh-CN', {year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined, month: '2-digit', day: '2-digit'});
    }
    function formatSize(bytes) {
        return bytes >= 1024 * 1024 ? (bytes / 1048576).toFixed(1) + ' MB' : (bytes / 1024).toFixed(1) + ' KB';
    }

    function showConfirm(message) {
        const modal = $('confirm-modal');
        $('confirm-msg').textContent = message;
        modal.returnValue = '';
        $('confirm-ok').onclick = () => modal.close('confirm');
        $('confirm-cancel').onclick = () => modal.close('cancel');
        modal.showModal();
        $('confirm-cancel').focus();
        return new Promise(resolve => modal.addEventListener('close', () => resolve(modal.returnValue === 'confirm'), {once: true}));
    }

    async function deleteReport(id, title) {
        if (!await showConfirm('确定删除「' + title + '」？删除后无法恢复。')) return;
        try {
            await request('/api/delete/' + encodeURIComponent(id), {method: 'DELETE', headers: authHeaders()});
            allReports = allReports.filter(r => r.id !== id);
            categories = [...new Set(allReports.map(r => r.category))];
            renderFilters(); render(); searchInput.focus(); notify('报告已删除');
        } catch (err) { notify(err.message); }
    }

    function reportURL(report) {
        // Report links are same-origin; never send a token to a metadata-supplied host.
        const url = new URL(report.url, window.location.origin);
        if (url.origin !== window.location.origin || !['http:', 'https:'].includes(url.protocol)) return '#';
        if (token) url.searchParams.set('token', token);
        return url.pathname + url.search;
    }

    function render() {
        if (!loaded) return;
        syncFilters();
        const filtered = getFiltered();
        $('result-count').textContent = hasFilters() ? '找到 ' + filtered.length + ' 份报告，共 ' + allReports.length + ' 份' : '共 ' + allReports.length + ' 份报告';
        emptyEl.style.display = filtered.length ? 'none' : 'block';
        if (!filtered.length) {
            content.innerHTML = '';
            $('empty-title').textContent = hasFilters() ? '没有找到匹配的报告' : '报告库还是空的';
            $('empty-description').textContent = hasFilters() ? '试试其他关键词，或清除筛选查看全部报告。' : user ? '上传一份 HTML 报告，开始建立你的资料库。' : '当前没有公开报告，登录后可查看自己的报告。';
            $('empty-action').textContent = hasFilters() ? '清除筛选' : user ? '上传报告' : '前往登录';
            return;
        }
        content.innerHTML = Object.entries(groupByDate(filtered)).filter(([, reports]) => reports.length).map(([label, reports]) =>
            '<section class="group"><h2 class="group-title">' + label + '<span class="group-count">' + reports.length + ' 份</span></h2><div class="cards">' +
            reports.map(r => {
                const canManage = user && (user.role === 'admin' || user.id === r.owner);
                const isPrivate = r.visibility === 'private';
                const visibility = isPrivate ? '私有' : '公开';
                const visAttrs = canManage ? 'button type="button"' : 'span';
                const visEnd = canManage ? 'button' : 'span';
                const visLabel = canManage ? '当前' + visibility + '，点击切换为' + (isPrivate ? '公开' : '私有') : visibility + '报告';
                const vis = '<' + visAttrs + ' class="vis-toggle' + (canManage ? ' clickable' : '') + '" data-id="' + escapeHtml(r.id) + '" data-vis="' + escapeHtml(r.visibility) + '" title="' + visLabel + '" aria-label="' + visLabel + '">' + icon(isPrivate ? 'lock' : 'globe') + '</' + visEnd + '>';
                return '<article class="card"><div class="card-heading">' +
                    '<h3 class="card-title"><a class="card-link" href="' + escapeHtml(reportURL(r)) + '" title="' + escapeHtml(r.title) + '" target="_blank" rel="noopener noreferrer"><span class="card-title-text">' + escapeHtml(r.title) + '</span></a></h3>' + vis + '</div>' +
                    (r.summary ? '<p class="card-summary">' + escapeHtml(r.summary) + '</p>' : '') +
                    '<div class="card-meta"><span class="badge">' + escapeHtml(r.category) + '</span>' + (r.tags || []).map(tag => '<button class="tag ' + tagColorClass(tag) + '" data-tag="' + escapeHtml(tag) + '" aria-label="筛选标签：' + escapeHtml(tag) + '">' + escapeHtml(tag) + '</button>').join('') + '</div>' +
                    '<div class="card-footer"><time datetime="' + escapeHtml(r.created_at) + '" title="' + escapeHtml(new Date(r.created_at).toLocaleString('zh-CN')) + '">' + formatTime(r.created_at) + '</time><span>' + formatSize(r.size) + '</span><span class="card-owner">' + escapeHtml(ownerMap[r.owner] || '') + '</span>' +
                    (canManage ? '<button class="card-delete" data-id="' + escapeHtml(r.id) + '" data-title="' + escapeHtml(r.title) + '" aria-label="删除报告：' + escapeHtml(r.title) + '">' + icon('trash') + '</button>' : '') + '</div></article>';
            }).join('') + '</div></section>'
        ).join('');
    }

    content.addEventListener('click', async e => {
        const tag = e.target.closest('button[data-tag]');
        if (tag) {
            selectedTag = selectedTag === tag.dataset.tag ? '' : tag.dataset.tag;
            render();
            const selected = isDateTag(selectedTag) ? $('date-tag-select') : [...tagButtons.querySelectorAll('button')].find(button => button.dataset.tag === selectedTag);
            if (selected) selected.focus({preventScroll: true});
            return;
        }
        const del = e.target.closest('.card-delete');
        if (del) { deleteReport(del.dataset.id, del.dataset.title); return; }
        const vis = e.target.closest('.vis-toggle.clickable');
        if (!vis) return;
        vis.disabled = true;
        const next = vis.dataset.vis === 'private' ? 'public' : 'private';
        try {
            await request('/api/report/' + encodeURIComponent(vis.dataset.id) + '/visibility', {
                method: 'PUT', headers: {...authHeaders(), 'Content-Type': 'application/json'},
                body: JSON.stringify({visibility: next})
            });
            const report = allReports.find(r => r.id === vis.dataset.id);
            if (report) report.visibility = next;
            const reportId = vis.dataset.id;
            render();
            const replacement = [...content.querySelectorAll('.vis-toggle.clickable')].find(button => button.dataset.id === reportId);
            if (replacement) replacement.focus();
            notify(next === 'private' ? '已设为私有报告' : '已设为公开报告');
        } catch (err) { vis.disabled = false; notify(err.message); }
    });

    function escapeHtml(value) {
        return String(value ?? '').replace(/[&<>"']/g, char => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[char]));
    }

    let timer;
    searchInput.addEventListener('input', () => {
        clearTimeout(timer); timer = setTimeout(render, 180);
    });
    init();
})();
