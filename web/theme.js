// Apply the saved theme before styles load on every page.
(function () {
    try {
        const saved = localStorage.getItem('theme');
        const dark = saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches);
        document.documentElement.dataset.theme = dark ? 'dark' : 'light';
    } catch (_) { /* Storage may be unavailable in private browsing. */ }
})();
