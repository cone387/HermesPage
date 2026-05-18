(function() {
    const content = document.getElementById('content');
    const emptyEl = document.getElementById('empty');
    const searchInput = document.getElementById('search');
    const categoryButtons = document.getElementById('category-buttons');
    const tagButtons = document.getElementById('tag-buttons');
    const ownerButtons = document.getElementById('owner-buttons');
    const ownerFilterRow = document.getElementById('owner-filter-row');
    const userInfoEl = document.getElementById('user-info');
    const userMenu = document.getElementById('user-menu');
    const userMenuTrigger = document.getElementById('user-menu-trigger');
    const userDropdown = document.getElementById('user-dropdown');
    const logoutBtn = document.getElementById('logout-btn');
    const sortSelect = document.getElementById('sort-select');
    const themeToggle = document.getElementById('theme-toggle');
    const uploadBtn = document.getElementById('upload-btn');
    const uploadModal = document.getElementById('upload-modal');

    let allReports = [];
    let categories = [];
    let ownerMap = {};
    let selectedCategory = '';
    let selectedTag = '';
    let selectedOwner = '';
    let sortBy = 'time-desc';

    const token = localStorage.getItem('token');
    const user = JSON.parse(localStorage.getItem('user') || 'null');

    // dark mode - prefer saved, fallback to system preference
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
        document.documentElement.setAttribute('data-theme', 'dark');
    }

    async function init() {
        try {
            const statusResp = await fetch('/api/setup/status');
            const status = await statusResp.json();
            if (status.needs_setup) {
                window.location.href = '/setup.html';
                return;
            }
        } catch (e) {}

        if (user) {
            userMenu.style.display = 'block';
            userMenuTrigger.textContent = user.username;
            userInfoEl.style.display = 'none';
            if (uploadBtn) uploadBtn.style.display = 'inline-block';

            userMenuTrigger.addEventListener('click', (e) => {
                e.stopPropagation();
                userDropdown.classList.toggle('show');
            });
            document.addEventListener('click', () => {
                userDropdown.classList.remove('show');
            });
        } else {
            userInfoEl.innerHTML = '<a href="/login.html" style="color:var(--primary);text-decoration:none;font-size:0.82rem">登录</a>';
        }

        logoutBtn.addEventListener('click', () => {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.reload();
        });

        if (themeToggle) {
            themeToggle.addEventListener('click', () => {
                const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
                if (isDark) {
                    document.documentElement.removeAttribute('data-theme');
                    localStorage.removeItem('theme');
                } else {
                    document.documentElement.setAttribute('data-theme', 'dark');
                    localStorage.setItem('theme', 'dark');
                }
            });
        }

        if (sortSelect) {
            sortSelect.addEventListener('change', () => {
                sortBy = sortSelect.value;
                render();
            });
        }

        if (uploadBtn) {
            uploadBtn.addEventListener('click', () => {
                uploadModal.classList.add('show');
            });
        }

        initUpload();
        await fetchReports();
    }

    function initUpload() {
        if (!uploadModal) return;
        const closeBtn = uploadModal.querySelector('.modal-close');
        const dropZone = document.getElementById('drop-zone');
        const fileInput = document.getElementById('file-input');
        const uploadForm = document.getElementById('upload-form');

        closeBtn.addEventListener('click', () => uploadModal.classList.remove('show'));
        uploadModal.addEventListener('click', (e) => {
            if (e.target === uploadModal) uploadModal.classList.remove('show');
        });

        dropZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            dropZone.classList.add('dragover');
        });
        dropZone.addEventListener('dragleave', () => dropZone.classList.remove('dragover'));
        dropZone.addEventListener('drop', (e) => {
            e.preventDefault();
            dropZone.classList.remove('dragover');
            if (e.dataTransfer.files.length > 0) {
                fileInput.files = e.dataTransfer.files;
                dropZone.querySelector('.drop-text').textContent = e.dataTransfer.files[0].name;
            }
        });
        dropZone.addEventListener('click', () => fileInput.click());
        fileInput.addEventListener('change', () => {
            if (fileInput.files.length > 0) {
                dropZone.querySelector('.drop-text').textContent = fileInput.files[0].name;
            }
        });

        uploadForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (!fileInput.files.length) return;

            const formData = new FormData();
            formData.append('file', fileInput.files[0]);
            const title = document.getElementById('upload-title').value;
            const category = document.getElementById('upload-category').value;
            const tags = document.getElementById('upload-tags').value;
            const visibility = document.getElementById('upload-visibility').value;
            if (title) formData.append('title', title);
            if (category) formData.append('category', category);
            if (tags) formData.append('tags', tags);
            formData.append('visibility', visibility);

            const resp = await fetch('/api/upload', {
                method: 'POST',
                headers: { 'Authorization': 'Bearer ' + token },
                body: formData
            });
            if (resp.ok) {
                uploadModal.classList.remove('show');
                uploadForm.reset();
                dropZone.querySelector('.drop-text').textContent = '拖拽 HTML 文件到此处，或点击选择';
                await fetchReports();
            } else {
                const data = await resp.json();
                alert(data.error || '上传失败');
            }
        });
    }

    function authHeaders() {
        const headers = {};
        if (token) headers['Authorization'] = 'Bearer ' + token;
        return headers;
    }

    async function fetchReports() {
        try {
            const resp = await fetch('/api/list', { headers: authHeaders() });
            const data = await resp.json();
            allReports = data.reports || [];
            categories = data.categories || [];
            ownerMap = data.owners || {};
            renderFilters();
            render();
        } catch (e) {
            content.innerHTML = '<div class="loading">加载失败，请刷新重试</div>';
        }
    }

    function getAllTags() {
        const tagSet = new Set();
        allReports.forEach(r => (r.tags || []).forEach(t => tagSet.add(t)));
        return Array.from(tagSet).sort();
    }

    function tagColorClass(tag) {
        let hash = 0;
        for (let i = 0; i < tag.length; i++) {
            hash = tag.charCodeAt(i) + ((hash << 5) - hash);
        }
        return 'tag-' + (Math.abs(hash) % 6);
    }

    function renderFilters() {
        let catHtml = '<button class="cat-btn active" data-cat="">全部</button>';
        categories.sort().forEach(cat => {
            catHtml += `<button class="cat-btn" data-cat="${cat}">${cat}</button>`;
        });
        categoryButtons.innerHTML = catHtml;

        const allTags = getAllTags();
        let tagHtml = '<button class="tag-btn active" data-tag="">全部</button>';
        allTags.forEach(tag => {
            tagHtml += `<button class="tag-btn ${tagColorClass(tag)}" data-tag="${tag}">${tag}</button>`;
        });
        tagButtons.innerHTML = tagHtml;

        if (user && user.role === 'admin' && Object.keys(ownerMap).length > 0) {
            ownerFilterRow.style.display = 'flex';
            let ownerHtml = '<button class="cat-btn active" data-owner="">全部</button>';
            Object.entries(ownerMap).forEach(([id, name]) => {
                ownerHtml += `<button class="cat-btn" data-owner="${id}">${name}</button>`;
            });
            ownerButtons.innerHTML = ownerHtml;

            ownerButtons.querySelectorAll('.cat-btn').forEach(btn => {
                btn.addEventListener('click', () => {
                    selectedOwner = btn.dataset.owner;
                    ownerButtons.querySelectorAll('.cat-btn').forEach(b => b.classList.remove('active'));
                    btn.classList.add('active');
                    render();
                });
            });
        }

        categoryButtons.querySelectorAll('.cat-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                selectedCategory = btn.dataset.cat;
                categoryButtons.querySelectorAll('.cat-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                render();
            });
        });

        tagButtons.querySelectorAll('.tag-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                selectedTag = btn.dataset.tag;
                tagButtons.querySelectorAll('.tag-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                render();
            });
        });
    }

    function getFiltered() {
        const search = searchInput.value.toLowerCase().trim();
        let filtered = allReports.filter(r => {
            if (selectedCategory && r.category !== selectedCategory) return false;
            if (selectedTag && !(r.tags || []).includes(selectedTag)) return false;
            if (selectedOwner && r.owner !== selectedOwner) return false;
            if (search) {
                const inTitle = r.title.toLowerCase().includes(search);
                const inTags = (r.tags || []).some(t => t.toLowerCase().includes(search));
                if (!inTitle && !inTags) return false;
            }
            return true;
        });

        // sort
        filtered.sort((a, b) => {
            switch (sortBy) {
                case 'time-asc': return new Date(a.created_at) - new Date(b.created_at);
                case 'title-asc': return a.title.localeCompare(b.title);
                case 'title-desc': return b.title.localeCompare(a.title);
                case 'size-desc': return b.size - a.size;
                case 'size-asc': return a.size - b.size;
                default: return new Date(b.created_at) - new Date(a.created_at);
            }
        });

        return filtered;
    }

    function groupByDate(reports) {
        if (sortBy !== 'time-desc' && sortBy !== 'time-asc') {
            return { '全部': reports };
        }
        const now = new Date();
        const today = dateKey(now);
        const yesterday = dateKey(new Date(now - 86400000));
        const weekAgo = new Date(now - 7 * 86400000);

        const groups = { '今天': [], '昨天': [], '本周': [], '更早': [] };

        reports.forEach(r => {
            const d = new Date(r.created_at);
            const key = dateKey(d);
            if (key === today) groups['今天'].push(r);
            else if (key === yesterday) groups['昨天'].push(r);
            else if (d > weekAgo) groups['本周'].push(r);
            else groups['更早'].push(r);
        });

        return groups;
    }

    function dateKey(d) {
        return d.getFullYear() + '-' + String(d.getMonth()+1).padStart(2,'0') + '-' + String(d.getDate()).padStart(2,'0');
    }

    function formatTime(dateStr) {
        const d = new Date(dateStr);
        const now = new Date();
        if (dateKey(d) === dateKey(now)) {
            return String(d.getHours()).padStart(2,'0') + ':' + String(d.getMinutes()).padStart(2,'0');
        }
        return (d.getMonth()+1) + '/' + d.getDate() + ' ' + String(d.getHours()).padStart(2,'0') + ':' + String(d.getMinutes()).padStart(2,'0');
    }

    function formatSize(bytes) {
        if (bytes > 1024 * 1024) return (bytes / (1024*1024)).toFixed(1) + ' MB';
        return (bytes / 1024).toFixed(1) + ' KB';
    }

    function showConfirm(msg) {
        return new Promise((resolve) => {
            const modal = document.getElementById('confirm-modal');
            document.getElementById('confirm-msg').textContent = msg;
            modal.classList.add('show');
            const ok = document.getElementById('confirm-ok');
            const cancel = document.getElementById('confirm-cancel');
            function cleanup() {
                modal.classList.remove('show');
                ok.replaceWith(ok.cloneNode(true));
                cancel.replaceWith(cancel.cloneNode(true));
            }
            document.getElementById('confirm-ok').addEventListener('click', () => { cleanup(); resolve(true); });
            document.getElementById('confirm-cancel').addEventListener('click', () => { cleanup(); resolve(false); });
        });
    }

    async function deleteReport(id, title) {
        const yes = await showConfirm(`确定删除「${title}」？此操作不可撤销。`);
        if (!yes) return;
        const resp = await fetch(`/api/delete/${id}`, {
            method: 'DELETE',
            headers: authHeaders()
        });
        if (resp.ok) {
            allReports = allReports.filter(r => r.id !== id);
            renderFilters();
            render();
        }
    }

    function render() {
        const filtered = getFiltered();
        if (filtered.length === 0) {
            content.innerHTML = '';
            emptyEl.style.display = 'block';
            return;
        }
        emptyEl.style.display = 'none';

        const groups = groupByDate(filtered);
        let html = '';

        for (const [label, reports] of Object.entries(groups)) {
            if (reports.length === 0) continue;
            html += `<div class="group">`;
            html += `<div class="group-title">${label}</div>`;
            html += `<div class="cards">`;
            reports.forEach(r => {
                const tagsHtml = (r.tags || []).map(t =>
                    `<span class="tag ${tagColorClass(t)}" data-tag="${t}">${t}</span>`
                ).join('');
                const canToggle = user && (user.role === 'admin' || user.id === r.owner);
                const visHtml = r.visibility === 'private'
                    ? `<span class="vis-toggle${canToggle ? ' clickable' : ''}" data-id="${r.id}" data-vis="private" title="私有（点击切换为公开）">🔒</span>`
                    : `<span class="vis-toggle${canToggle ? ' clickable' : ''}" data-id="${r.id}" data-vis="public" title="公开（点击切换为私有）">🌐</span>`;
                const ownerName = ownerMap[r.owner] || '';
                const canDelete = user && (user.role === 'admin' || user.id === r.owner);
                const deleteHtml = canDelete ? `<span class="card-delete" data-id="${r.id}" data-title="${escapeHtml(r.title)}" title="删除">×</span>` : '';
                html += `
                    <div class="card" data-url="${r.url}">
                        <div class="card-title"><span class="card-title-text">${escapeHtml(r.title)}</span>${visHtml}</div>
                        <div class="card-meta">
                            <span class="badge">${r.category}</span>
                            ${tagsHtml}
                        </div>
                        <div class="card-footer">
                            <span>${formatTime(r.created_at)}</span>
                            <span class="card-owner">${ownerName}</span>
                            <span>${formatSize(r.size)}</span>
                            ${deleteHtml}
                        </div>
                    </div>`;
            });
            html += `</div></div>`;
        }

        content.innerHTML = html;

        content.querySelectorAll('.card').forEach(card => {
            card.addEventListener('click', (e) => {
                if (e.target.classList.contains('tag')) return;
                if (e.target.classList.contains('vis-toggle')) return;
                if (e.target.classList.contains('card-delete')) return;
                const url = card.dataset.url;
                if (token) {
                    window.open(url + '?token=' + encodeURIComponent(token), '_blank');
                } else {
                    window.open(url, '_blank');
                }
            });
        });
        content.querySelectorAll('.card .tag').forEach(tagEl => {
            tagEl.addEventListener('click', (e) => {
                e.stopPropagation();
                const tag = tagEl.dataset.tag;
                selectedTag = (selectedTag === tag) ? '' : tag;
                tagButtons.querySelectorAll('.tag-btn').forEach(b => {
                    b.classList.toggle('active', b.dataset.tag === (selectedTag || ''));
                });
                render();
            });
        });
        content.querySelectorAll('.vis-toggle.clickable').forEach(el => {
            el.addEventListener('click', async (e) => {
                e.stopPropagation();
                const id = el.dataset.id;
                const newVis = el.dataset.vis === 'private' ? 'public' : 'private';
                const resp = await fetch(`/api/report/${id}/visibility`, {
                    method: 'PUT',
                    headers: { 'Authorization': 'Bearer ' + token, 'Content-Type': 'application/json' },
                    body: JSON.stringify({ visibility: newVis })
                });
                if (resp.ok) {
                    const idx = allReports.findIndex(r => r.id === id);
                    if (idx >= 0) allReports[idx].visibility = newVis;
                    render();
                }
            });
        });
        content.querySelectorAll('.card-delete').forEach(el => {
            el.addEventListener('click', (e) => {
                e.stopPropagation();
                deleteReport(el.dataset.id, el.dataset.title);
            });
        });
    }

    function escapeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }

    let timer;
    searchInput.addEventListener('input', () => {
        clearTimeout(timer);
        timer = setTimeout(render, 300);
    });

    init();
})();
