const {test} = require('node:test');
const assert = require('node:assert/strict');
const vm = require('node:vm');
const fs = require('node:fs');

// Exercise the actual application's pure data helpers without a browser dependency.
function app() {
    const element = {addEventListener() {}, value: '', style: {}, setAttribute() {}};
    const context = vm.createContext({
        document: {getElementById: () => element, documentElement: element, addEventListener() {},
            createElement: () => ({textContent: '', get innerHTML() {return this.textContent.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');}})},
        localStorage: {getItem: () => null},
        window: {matchMedia: () => ({matches: false})},
        fetch: async () => { throw new Error('offline'); },
        Date, URL, console, setTimeout, clearTimeout,
    });
    const source = fs.readFileSync('web/app.js', 'utf8').replace(/    init\(\);\s*\}\)\(\);\s*$/, `
    globalThis.helpers = {groupByDate, escapeHtml, getFiltered, fetchReports, partitionTags,
        isLoaded: () => loaded,
        setState(reports, sort) { allReports = reports; sortBy = sort; loaded = true; },
        setSearch(value) { searchInput.value = value; }};
})();`);
    vm.runInContext(source, context);
    return context.helpers;
}

test('oldest-first keeps date groups in chronological order', () => {
    const h = app();
    h.setState([], 'time-asc');
    const today = new Date();
    const groups = h.groupByDate([
        {created_at: new Date(today - 20 * 86400000).toISOString()},
        {created_at: today.toISOString()},
    ]);
    assert.deepEqual(Object.entries(groups).filter(([,v]) => v.length).map(([k]) => k), ['更早', '今天']);
});

test('metadata is safe in text and quoted HTML attributes', () => {
    assert.equal(app().escapeHtml('<img title="x" onerror=\'bad\'>&'), '&lt;img title=&quot;x&quot; onerror=&#39;bad&#39;&gt;&amp;');
});

test('failed refresh prevents subsequent filters from presenting stale reports', async () => {
    const h = app();
    h.setState([{title: 'Previously loaded'}], 'time-desc');
    await h.fetchReports();
    assert.equal(h.isLoaded(), false);
});

test('date tags are separated from topics and sorted newest first', () => {
    const tags = ['2026-01-02', 'AI', '2026-10-27', '2026-02-30', '2026 年报', 'release-2026-10-27'];
    const groups = app().partitionTags(tags);
    assert.deepEqual(Array.from(groups.dates), ['2026-10-27', '2026-01-02']);
    assert.deepEqual(Array.from(groups.topics), ['AI', '2026-02-30', '2026 年报', 'release-2026-10-27']);
    assert.equal(tags[0], '2026-01-02');
});

test('empty tag lists do not create a date group', () => {
    const groups = app().partitionTags([]);
    assert.equal(groups.dates.length, 0);
    assert.equal(groups.topics.length, 0);
});

test('search finds summary text and tolerates reports without summaries', () => {
    const h = app();
    h.setState([
        {id: 'excerpt', title: 'Daily report', summary: '跨会话检索与 Memory benchmarks', created_at: '2026-10-27'},
        {id: 'legacy', title: 'Older report', created_at: '2026-10-26'},
    ], 'time-desc');
    h.setSearch('  memory  ');
    assert.deepEqual(Array.from(h.getFiltered(), r => r.id), ['excerpt']);
    h.setSearch('跨会话');
    assert.deepEqual(Array.from(h.getFiltered(), r => r.id), ['excerpt']);
    h.setSearch('older');
    assert.deepEqual(Array.from(h.getFiltered(), r => r.id), ['legacy']);
});
