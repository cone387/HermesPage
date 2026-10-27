# Acta UI polish and report summaries

Goal: improve visual consistency while preserving the user's existing interface and navigation.

Preserve the original top search, horizontal category/tag/owner filters, sorting control, and date-grouped card grid. Category names stay unchanged. Keep topic tags directly visible, and place valid ISO calendar-date tags in one native select, newest first. The selected date remains visible in this control; topic/date selection retains the existing single-tag filter behavior.

Apply UI/UX Pro Max guidance for neutral gray/link-blue colors, contrast, typography, spacing, focus and responsive behavior within the existing layout.

Design: white top filter area, soft blue-gray page background, compact neutral filter chips with blue selected states. Cards show title, summary, category and colored tags, then metadata. Search, upload, theme and account controls remain in the header. Mobile wraps the header and filters and uses a single card column. Keep native dialogs and existing API operations.

Report excerpts: show escaped plain text below each card title, capped at three lines. The server extracts the first 120 Unicode characters from HTML text after dropping scripts, styles, navigation and main titles; empty excerpts add no placeholder. New uploads persist the excerpt; older records backfill it from their HTML once at startup without changing report files. No extra per-card browser requests. Go tests cover HTML cleanup, entities, Unicode truncation, persistence and old metadata/missing files. Desktop and 375px browser checks show 181 excerpts, a three-line height of 64.35px and no horizontal overflow.

Final review: homepage search now includes summary text with case-insensitive matching and handles older records without summaries (six Node tests total). Browser confirmed a summary-only keyword finds the expected report. Independent review found an omitted `</head>` compatibility issue; removing non-content head elements individually preserves body text, covered by two regression cases. Report visibility checks still precede summary serialization. README explains extraction limits and the required backend + static file update on deployment.

- [x] Restore original homepage structure, visible tags, raw category names and title-first cards.
- [x] Remove sidebar/collapsible-tag CSS and JavaScript dependencies.
- [x] Refine typography, tag colors, focus indicators, spacing and mobile wrapping.
- [x] Validate category/tag selection, reset, desktop/mobile/light/dark layout and regressions.

Validation: 181 synthetic local reports and 155 tags. Playwright verified search inside header, category and card-tag filtering, reset, 375px mobile width without page overflow, dark cards and long-title keyboard focus without clipping. The date-tag revision shows 12 topic tags plus All, with 143 date options; the desktop filter area shrank from 354px to 126px. Date selection, switching to a topic, card-date clicks and clearing filters were checked. Six Node tests cover oldest-first grouping, escaped metadata, failed-refresh state, valid date partitioning/latest-first ordering, empty tags and summary search. No deployment or production data changes. Preview fixtures and screenshots remain under ignored .playwright-mcp/ui-review/.
