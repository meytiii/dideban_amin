window.addEventListener('error', (e) => {
    console.error('[GLOBAL ERROR]', e.error || e.message, e.filename, e.lineno);
});

window.addEventListener('unhandledrejection', (e) => {
    console.error('[UNHANDLED PROMISE]', e.reason);
});

function fetchWithTimeout(url, options = {}, ms = 20000) {
    const ctrl = new AbortController();
    const timer = setTimeout(() => ctrl.abort(), ms);
    return fetch(url, { ...options, signal: ctrl.signal })
        .catch(err => {
            if (url !== '/health' && typeof window.__checkAdminHostConnection === 'function') {
                window.__checkAdminHostConnection();
            }
            throw err;
        })
        .finally(() => clearTimeout(timer));
}

(function initAdminPreloader() {
    let dismissed = false;

    function dismissAdminPreloader() {
        if (dismissed) return;
        dismissed = true;
        const preloader = document.getElementById('appPreloader') || document.getElementById('adminPreloader');
        const skeleton = document.getElementById('pageSkeleton') || document.getElementById('adminSkeleton');
        if (preloader) {
            preloader.classList.add('fade-out');
            setTimeout(() => preloader.remove(), 450);
        }
        if (skeleton) skeleton.remove();
    }

    window.markAdminReady = function () {
        dismissAdminPreloader();
    };

    function onReadyCheck() {
        if (window.__adminStartupComplete) window.markAdminReady();
    }

    if (document.readyState === 'complete') {
        onReadyCheck();
    } else {
        window.addEventListener('load', onReadyCheck);
        document.addEventListener('DOMContentLoaded', onReadyCheck);
    }

    setTimeout(() => {
        window.markAdminReady();
    }, 10000);
})();

function getCsrfToken() {
    const match = document.cookie.match(/(?:^|;\s*)herasat_csrf=([^;]+)/);
    if (match && match[1]) {
        const val = decodeURIComponent(match[1]);
        sessionStorage.setItem('herasat_csrf', val);
        return val;
    }
    return sessionStorage.getItem('herasat_csrf') || '';
}

function adminFetch(url, options = {}) {
    const opts = { credentials: 'same-origin', ...options };
    opts.headers = { ...(opts.headers || {}) };
    const csrf = getCsrfToken();
    if (csrf && opts.method && opts.method.toUpperCase() !== 'GET') {
        opts.headers['X-CSRF-Token'] = csrf;
    }
    const t0 = (typeof performance !== 'undefined' && performance.now) ? performance.now() : Date.now();
    return fetchWithTimeout(url, opts, 20000).then(async res => {
        const t1 = (typeof performance !== 'undefined' && performance.now) ? performance.now() : Date.now();
        const rtt = Math.max(0, t1 - t0);
        const serverHdr = res.headers && (res.headers.get('x-server-time') || res.headers.get('date'));
        if (serverHdr && typeof window.__syncServerTime === 'function') {
            window.__syncServerTime(serverHdr, rtt);
        }
        const newCsrf = res.headers && res.headers.get('x-csrf-token');
        if (newCsrf) {
            sessionStorage.setItem('herasat_csrf', newCsrf);
        }
        if (res.status === 401) {
            if (typeof forceLogoutOnConnectionDrop === 'function') {
                forceLogoutOnConnectionDrop();
            }
        }
        if (res.status === 403 && opts.method && opts.method.toUpperCase() !== 'GET' && !opts._retried) {
            const clone = res.clone();
            const text = await clone.text().catch(() => '');
            if (text.includes('CSRF')) {
                try {
                    const meRes = await fetchWithTimeout('/api/admin/me', { credentials: 'same-origin' }, 10000);
                    if (meRes.ok) {
                        const meData = await meRes.json();
                        const refreshedCsrf = meData.csrf_token || getCsrfToken();
                        if (refreshedCsrf) {
                            sessionStorage.setItem('herasat_csrf', refreshedCsrf);
                            opts._retried = true;
                            opts.headers['X-CSRF-Token'] = refreshedCsrf;
                            return fetchWithTimeout(url, opts, 20000);
                        }
                    }
                } catch (_) {}
            }
        }
        return res;
    });
}

let serverBaseMs = 0;
let perfBase = 0;
let serverTimeOffsetMs = 0;
let hasServerTimeSync = false;
let serverTimeZone = 'Asia/Tehran';

window.__syncServerTime = function (serverTimeInput, rttMs = 0, tz) {
    if (!serverTimeInput) return;
    try {
        let serverMs = 0;
        if (typeof serverTimeInput === 'number') {
            serverMs = serverTimeInput;
        } else if (typeof serverTimeInput === 'string') {
            const parsed = new Date(serverTimeInput).getTime();
            if (!isNaN(parsed) && parsed > 0) {
                serverMs = parsed;
            }
        }
        if (serverMs > 0) {
            const halfRtt = (typeof rttMs === 'number' && !isNaN(rttMs) && rttMs >= 0) ? Math.min(Math.round(rttMs / 2), 5000) : 0;
            const adjustedServerMs = serverMs + halfRtt;
            const nowPerf = (typeof performance !== 'undefined' && performance.now) ? performance.now() : 0;

            serverBaseMs = adjustedServerMs;
            perfBase = nowPerf;
            serverTimeOffsetMs = adjustedServerMs - Date.now();
            hasServerTimeSync = true;
            if (tz && typeof tz === 'string') {
                serverTimeZone = tz;
            }
            if (typeof window.__onServerTimeUpdated === 'function') {
                window.__onServerTimeUpdated();
            }
        }
    } catch (_) { }
};

function getServerNow() {
    if (hasServerTimeSync && serverBaseMs > 0 && typeof performance !== 'undefined' && performance.now) {
        return new Date(serverBaseMs + (performance.now() - perfBase));
    }
    return new Date(Date.now() + serverTimeOffsetMs);
}
window.getServerNow = getServerNow;

async function fetchServerTime() {
    try {
        const t0 = (typeof performance !== 'undefined' && performance.now) ? performance.now() : Date.now();
        const res = await fetch('/api/server-time', {
            method: 'GET',
            cache: 'no-store',
            credentials: 'same-origin'
        });
        const t1 = (typeof performance !== 'undefined' && performance.now) ? performance.now() : Date.now();
        if (res.ok) {
            const data = await res.json();
            if (data && (data.server_time_unix_ms || data.server_time)) {
                window.__syncServerTime(data.server_time_unix_ms || data.server_time, Math.max(0, t1 - t0), data.timezone);
            }
        }
    } catch (_) { }
}
window.fetchServerTime = fetchServerTime;
fetchServerTime();
setInterval(fetchServerTime, 60000);
document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') {
        fetchServerTime();
    }
});

const getScrollBehavior = () => (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth');

function promoteTemporarily(el, props = 'transform, opacity', durationMs = 300) {
    if (!el || !el.style) return;
    el.style.willChange = props;
    setTimeout(() => {
        if (el && el.style) {
            el.style.willChange = 'auto';
        }
    }, durationMs);
}

const runWhenIdle = (typeof window !== 'undefined' && window.requestIdleCallback)
    ? (cb) => window.requestIdleCallback(cb, { timeout: 1500 })
    : (cb) => setTimeout(cb, 100);

function initAppTooltips() {
    let tooltipEl = null;
    let activeTarget = null;
    let showTimer = null;
    let hideTimer = null;

    function getOrCreateTooltip() {
        if (!tooltipEl) {
            tooltipEl = document.createElement('div');
            tooltipEl.className = 'app-tooltip';
            tooltipEl.setAttribute('role', 'tooltip');
            tooltipEl.innerHTML = '<span class="app-tooltip-text"></span><div class="app-tooltip-arrow"></div>';
            document.body.appendChild(tooltipEl);
        }
        return tooltipEl;
    }

    function convertTitles(root) {
        if (!root || !root.querySelectorAll) return;
        root.querySelectorAll('[title]').forEach(el => {
            const t = el.getAttribute('title');
            if (t && t.trim()) {
                el.setAttribute('data-tooltip', t.trim());
                el.removeAttribute('title');
            }
        });
    }
    window.convertTitles = convertTitles;

    function isMobileOrTouch() {
        if (window.innerWidth <= 768) return true;
        try {
            if (window.matchMedia && (window.matchMedia('(hover: none)').matches || window.matchMedia('(pointer: coarse)').matches)) {
                return true;
            }
        } catch (e) { }
        return false;
    }

    function positionTooltip(el) {
        if (isMobileOrTouch()) return;
        const text = el.getAttribute('data-tooltip');
        if (!text || !text.trim()) return;

        const tip = getOrCreateTooltip();
        tip.querySelector('.app-tooltip-text').textContent = text;

        const rect = el.getBoundingClientRect();
        if (rect.width === 0 && rect.height === 0) return;

        tip.style.left = '-9999px';
        tip.style.top = '-9999px';
        tip.style.visibility = 'hidden';
        tip.style.display = 'block';

        const tipWidth = tip.offsetWidth;
        const tipHeight = tip.offsetHeight;

        const side = rect.top < (tipHeight + 14) ? 'bottom' : 'top';
        tip.setAttribute('data-side', side);

        let left = rect.left + (rect.width / 2) - (tipWidth / 2);
        left = Math.max(12, Math.min(window.innerWidth - tipWidth - 12, left));

        let top = side === 'top' ? (rect.top - tipHeight - 6) : (rect.bottom + 6);

        const arrow = tip.querySelector('.app-tooltip-arrow');
        if (arrow) {
            const arrowLeft = rect.left + (rect.width / 2) - left - 3;
            arrow.style.left = `${Math.max(8, Math.min(tipWidth - 14, arrowLeft))}px`;
        }

        tip.style.left = `${Math.round(left)}px`;
        tip.style.top = `${Math.round(top)}px`;
        tip.style.visibility = 'visible';

        requestAnimationFrame(() => {
            if (activeTarget === el && !isMobileOrTouch()) {
                tip.classList.add('show');
            }
        });
    }

    function show(el) {
        if (isMobileOrTouch()) return;
        if (activeTarget === el) return;
        activeTarget = el;
        clearTimeout(showTimer);
        clearTimeout(hideTimer);

        showTimer = setTimeout(() => {
            if (isMobileOrTouch()) {
                hide(el);
                return;
            }
            if (activeTarget === el) {
                positionTooltip(el);
            }
        }, 80);
    }

    function hide(el) {
        if (activeTarget !== el && el !== null) return;
        clearTimeout(showTimer);
        activeTarget = null;
        if (tooltipEl) {
            tooltipEl.classList.remove('show');
            hideTimer = setTimeout(() => {
                if (!activeTarget && tooltipEl) {
                    tooltipEl.style.display = 'none';
                    tooltipEl.style.visibility = 'hidden';
                }
            }, 160);
        }
    }

    const start = () => {
        convertTitles(document.body);
    };

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', start);
    } else {
        start();
    }

    document.addEventListener('pointerover', (e) => {
        if (e.pointerType === 'touch' || e.pointerType === 'pen') return;
        if (isMobileOrTouch()) return;
        const target = e.target.closest('[data-tooltip], [title]');
        if (target) {
            if (target.hasAttribute('title')) {
                const t = target.getAttribute('title');
                target.setAttribute('data-tooltip', t);
                target.removeAttribute('title');
            }
            show(target);
        }
    }, { passive: true });

    document.addEventListener('pointerout', (e) => {
        const target = e.target.closest('[data-tooltip]');
        if (target) {
            if (e.relatedTarget && target.contains(e.relatedTarget)) {
                return;
            }
            hide(target);
        }
    }, { passive: true });

    document.addEventListener('focusin', (e) => {
        if (isMobileOrTouch()) return;
        const target = e.target.closest('[data-tooltip], [title]');
        if (target) {
            if (target.hasAttribute('title')) {
                const t = target.getAttribute('title');
                target.setAttribute('data-tooltip', t);
                target.removeAttribute('title');
            }
            show(target);
        }
    }, { passive: true });

    document.addEventListener('focusout', (e) => {
        const target = e.target.closest('[data-tooltip]');
        if (target) hide(target);
    }, { passive: true });

    document.addEventListener('input', (e) => {
        if (activeTarget && (activeTarget === e.target || activeTarget.contains(e.target))) {
            hide(activeTarget);
        }
    }, { passive: true });

    document.addEventListener('touchstart', () => {
        if (activeTarget) hide(activeTarget);
        if (tooltipEl) {
            tooltipEl.classList.remove('show');
            tooltipEl.style.display = 'none';
            tooltipEl.style.visibility = 'hidden';
        }
    }, { passive: true });

    window.addEventListener('resize', () => {
        if (isMobileOrTouch()) {
            if (activeTarget) hide(activeTarget);
            if (tooltipEl) {
                tooltipEl.classList.remove('show');
                tooltipEl.style.display = 'none';
                tooltipEl.style.visibility = 'hidden';
            }
        }
    }, { passive: true });

    window.addEventListener('scroll', () => {
        if (activeTarget) hide(activeTarget);
    }, { passive: true });
}

try { runWhenIdle(initAppTooltips); } catch (e) { console.error('Tooltips init failed:', e); }

document.addEventListener('DOMContentLoaded', () => {

    const toPersianNum = (n) => String(n).replace(/[0-9]/g, d => '۰۱۲۳۴۵۶۷۸۹'[d]);

    function initServerSlidingClock() {
        const clockContainer = document.getElementById('serverSlidingClock');
        if (!clockContainer) return;

        const serverDateDayEl = document.getElementById('serverDateDay');
        const serverDateFullEl = document.getElementById('serverDateFull');

        const digitH1 = clockContainer.querySelector('[data-unit="h1"]');
        const digitH2 = clockContainer.querySelector('[data-unit="h2"]');
        const digitM1 = clockContainer.querySelector('[data-unit="m1"]');
        const digitM2 = clockContainer.querySelector('[data-unit="m2"]');
        const digitS1 = clockContainer.querySelector('[data-unit="s1"]');
        const digitS2 = clockContainer.querySelector('[data-unit="s2"]');

        const FA_DIGITS = ['۰', '۱', '۲', '۳', '۴', '۵', '۶', '۷', '۸', '۹'];
        const toFa = (n) => FA_DIGITS[n] || '۰';

        function updateDigit(container, newChar) {
            if (!container) return;
            const currentEl = container.querySelector('.digit-item:not(.slide-out)');
            if (!currentEl) {
                container.innerHTML = `<span class="digit-item">${newChar}</span>`;
                return;
            }
            if (currentEl.textContent === newChar) return;

            const oldSlides = container.querySelectorAll('.slide-out');
            oldSlides.forEach(el => el.remove());

            const nextEl = document.createElement('span');
            nextEl.className = 'digit-item slide-in-prep';
            nextEl.textContent = newChar;
            container.appendChild(nextEl);

            void nextEl.offsetWidth;

            nextEl.classList.remove('slide-in-prep');
            nextEl.classList.add('slide-in-active');
            currentEl.classList.add('slide-out');

            setTimeout(() => {
                if (currentEl.parentNode) currentEl.remove();
                nextEl.classList.remove('slide-in-active');
            }, 400);
        }

        let lastDayKey = '';

        const tehranTimeFormatter = new Intl.DateTimeFormat('en-US', {
            timeZone: 'Asia/Tehran',
            hour12: false,
            hourCycle: 'h23',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });

        const persianDateFormatter = new Intl.DateTimeFormat('fa-IR-u-ca-persian', {
            timeZone: 'Asia/Tehran',
            year: 'numeric',
            month: 'long',
            day: 'numeric',
            weekday: 'long'
        });

        function updateDate(now) {
            try {
                const parts = persianDateFormatter.formatToParts(now);
                let weekday = '';
                let day = '';
                let month = '';
                let year = '';
                parts.forEach(p => {
                    if (p.type === 'weekday') weekday = p.value;
                    if (p.type === 'day') day = p.value;
                    if (p.type === 'month') month = p.value;
                    if (p.type === 'year') year = p.value;
                });
                const dayKey = `${year}-${month}-${day}`;
                if (dayKey === lastDayKey) return;
                lastDayKey = dayKey;

                if (serverDateDayEl) serverDateDayEl.textContent = weekday ? `${weekday}،` : '';
                if (serverDateFullEl) serverDateFullEl.textContent = `${day} ${month} ${year}`;
            } catch (_) {
                if (serverDateFullEl) serverDateFullEl.textContent = now.toLocaleDateString('fa-IR');
            }
        }

        function tick() {
            const now = getServerNow();
            updateDate(now);

            let h = '00', m = '00', s = '00';
            try {
                const parts = tehranTimeFormatter.formatToParts(now);
                for (let i = 0; i < parts.length; i++) {
                    const p = parts[i];
                    if (p.type === 'hour') h = p.value;
                    else if (p.type === 'minute') m = p.value;
                    else if (p.type === 'second') s = p.value;
                }
            } catch (_) {
                h = String(now.getHours()).padStart(2, '0');
                m = String(now.getMinutes()).padStart(2, '0');
                s = String(now.getSeconds()).padStart(2, '0');
            }

            if (h.length === 1) h = '0' + h;
            if (m.length === 1) m = '0' + m;
            if (s.length === 1) s = '0' + s;

            updateDigit(digitH1, toFa(parseInt(h[0], 10)));
            updateDigit(digitH2, toFa(parseInt(h[1], 10)));
            updateDigit(digitM1, toFa(parseInt(m[0], 10)));
            updateDigit(digitM2, toFa(parseInt(m[1], 10)));
            updateDigit(digitS1, toFa(parseInt(s[0], 10)));
            updateDigit(digitS2, toFa(parseInt(s[1], 10)));
        }

        window.__onServerTimeUpdated = function () {
            tick();
        };

        tick();
        fetchServerTime().then(() => tick());
        setInterval(tick, 1000);
    }

    try {
        initServerSlidingClock();
    } catch (_) { }

    const toastContainer = document.getElementById('toast-container');
    const TOAST_ICONS = {
        success: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><path d="m9 11 3 3L22 4"/></svg>',
        error: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>',
        destructive: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>',
        warning: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>',
        info: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>',
        default: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>'
    };

    const TOAST_TITLES = {
        success: 'عملیات موفق',
        error: 'خطا در سامانه',
        destructive: 'خطا در سامانه',
        warning: 'توجه و هشدار',
        info: 'اطلاعیه سامانه',
        default: 'پیام سیستم'
    };

    function showToast(param1, param2 = 2800, param3 = '') {
        if (!toastContainer) return;
        let title = '';
        let description = '';
        let type = 'info';
        let duration = 2800;

        if (typeof param1 === 'object' && param1 !== null) {
            title = param1.title || '';
            description = param1.description || param1.message || '';
            type = param1.type || param1.variant || 'info';
            duration = typeof param1.duration === 'number' ? param1.duration : 2800;
        } else {
            description = String(param1 || '');
            if (typeof param2 === 'string') {
                type = param2 || 'info';
                duration = typeof param3 === 'number' ? param3 : 2800;
            } else if (typeof param2 === 'number') {
                duration = param2;
                type = typeof param3 === 'string' && param3 ? param3 : 'info';
            }
        }

        if (type === 'error') type = 'destructive';
        if (!type) type = 'info';
        if (!title) {
            title = TOAST_TITLES[type] || TOAST_TITLES.default;
        }

        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.setAttribute('role', 'alert');
        toast.setAttribute('aria-live', 'assertive');

        const iconSvg = TOAST_ICONS[type] || TOAST_ICONS.info;

        toast.innerHTML = `
            <span class="toast-icon" aria-hidden="true">${iconSvg}</span>
            <div class="toast-content">
                <h5 class="toast-title"></h5>
                <div class="toast-description"></div>
            </div>
            <button type="button" class="toast-close" aria-label="بستن">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
            </button>
        `;
        toast.querySelector('.toast-title').textContent = title;
        toast.querySelector('.toast-description').textContent = description;
        toast.querySelector('.toast-close').addEventListener('click', () => dismissToast(toast));
        toastContainer.appendChild(toast);
        if (duration > 0) {
            setTimeout(() => dismissToast(toast), duration);
        }
        return toast;
    }

    function dismissToast(toast) {
        if (!toast || !toast.parentNode) return;
        toast.classList.add('toast-exit');
        setTimeout(() => {
            if (toast && toast.parentNode) toast.parentNode.removeChild(toast);
        }, 260);
    }

    window.showToast = showToast;
    window.showIconAlert = showToast;

    document.addEventListener('invalid', function (e) {
        e.preventDefault();
    }, true);

    function getElementPersianLabel(element, formEl) {
        if (!element) return 'فیلد';
        if (element.id) {
            const label = (formEl || document).querySelector(`label[for="${element.id}"]`);
            if (label) {
                const text = label.textContent.replace('*', '').replace('(اختیاری)', '').replace(':', '').trim();
                if (text) return text;
            }
        }
        const parentLabel = element.closest('label');
        if (parentLabel) {
            const text = parentLabel.textContent.replace('*', '').replace('(اختیاری)', '').replace(':', '').trim();
            if (text) return text;
        }
        const ariaLabel = element.getAttribute('aria-label');
        if (ariaLabel && ariaLabel.trim()) return ariaLabel.trim();
        const placeholder = element.getAttribute('placeholder');
        if (placeholder && placeholder.trim()) {
            return placeholder.replace('مثال:', '').replace('...', '').trim();
        }
        return 'مورد نظر';
    }

    function highlightAndFocusInvalidField(field, message, formEl) {
        if (!field) return;
        const target = field.closest('.password-wrapper') || field.closest('.form-group') || field;
        target.classList.add('input-error', 'input-error-shake');
        setTimeout(() => target.classList.remove('input-error-shake'), 450);
        field.focus();
        if (typeof field.scrollIntoView === 'function') {
            field.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
        const clearHandler = function () {
            target.classList.remove('input-error');
            field.removeEventListener('input', clearHandler);
            field.removeEventListener('change', clearHandler);
        };
        field.addEventListener('input', clearHandler);
        field.addEventListener('change', clearHandler);
        showToast({
            title: 'تکمیل فیلدهای الزامی',
            description: message,
            type: 'destructive'
        });
    }

    document.addEventListener('submit', function (e) {
        const formEl = e.target;
        if (!formEl || typeof formEl.querySelectorAll !== 'function') return;

        const inputs = Array.from(formEl.querySelectorAll('input, select, textarea'));
        for (const input of inputs) {
            if (input.disabled || input.type === 'hidden' || input.type === 'button' || input.type === 'submit') continue;

            let isInvalid = false;
            let errorMsg = '';

            if (input.hasAttribute('required')) {
                if (input.type === 'checkbox' || input.type === 'radio') {
                    if (!input.checked) {
                        const name = input.name;
                        const groupChecked = name ? formEl.querySelector(`input[name="${name}"]:checked`) : null;
                        if (!groupChecked) {
                            isInvalid = true;
                            errorMsg = `لطفاً گزینه الزامی «${getElementPersianLabel(input, formEl)}» را انتخاب فرمایید.`;
                        }
                    }
                } else if (!input.value || !input.value.trim()) {
                    isInvalid = true;
                    errorMsg = `لطفاً فیلد «${getElementPersianLabel(input, formEl)}» را تکمیل فرمایید.`;
                }
            }

            if (!isInvalid && input.value && input.minLength > 0 && input.value.trim().length < input.minLength) {
                isInvalid = true;
                const minVal = typeof toPersianNum === 'function' ? toPersianNum(input.minLength) : input.minLength;
                errorMsg = `حداقل ${minVal} کاراکتر برای فیلد «${getElementPersianLabel(input, formEl)}» الزامی است.`;
            }

            if (!isInvalid && input.value && input.pattern) {
                try {
                    const regex = new RegExp(`^(?:${input.pattern})$`);
                    if (!regex.test(input.value.trim())) {
                        isInvalid = true;
                        errorMsg = `فرمت وارد شده برای فیلد «${getElementPersianLabel(input, formEl)}» معتبر نمی‌باشد.`;
                    }
                } catch (_) { }
            }

            if (isInvalid) {
                e.preventDefault();
                e.stopImmediatePropagation();
                highlightAndFocusInvalidField(input, errorMsg, formEl);
                return;
            }
        }
    }, true);

    document.querySelectorAll('input, select, textarea').forEach(element => {
        const clearHandler = function () {
            const target = element.closest('.password-wrapper') || element.closest('.form-group') || element;
            target.classList.remove('input-error');
        };
        element.addEventListener('input', clearHandler);
        element.addEventListener('change', clearHandler);
    });

    const showConfirmDialog = window.showConfirmDialog || function () { return Promise.resolve(true); };

    const loginForm = document.getElementById('loginForm');
    const loginSection = document.getElementById('login-section');
    const dashboardSection = document.getElementById('dashboard-section');
    const loginMessage = document.getElementById('loginMessage');
    const reportsBody = document.getElementById('reportsBody');

    const reportsInboxSection = document.getElementById('reports-inbox-section');
    const listSection = document.getElementById('reports-list-section');
    const detailSection = document.getElementById('report-detail-section');
    const filterContainer = document.getElementById('reportsFilterContainer') || document.querySelector('.filter-container');
    const bulkActionsToolbar = document.getElementById('bulkActionsToolbar');
    const advancedSection = document.getElementById('advanced-section');
    const analyticsSection = document.getElementById('analytics-section');
    const tabBtnReports = document.getElementById('tab-btn-reports');
    const tabBtnAnalytics = document.getElementById('tab-btn-analytics');
    const tabBtnAdvanced = document.getElementById('tab-btn-advanced');
    const backBtn = document.getElementById('backToListBtn');

    let currentAdmin = null;

    if (loginSection) loginSection.style.display = 'flex';
    if (dashboardSection) dashboardSection.style.display = 'none';
    if (reportsInboxSection) reportsInboxSection.style.display = 'block';
    if (detailSection) detailSection.style.display = 'none';
    if (listSection) listSection.style.display = 'block';
    if (filterContainer) filterContainer.style.display = '';
    if (advancedSection) advancedSection.style.display = 'none';
    if (analyticsSection) analyticsSection.style.display = 'none';

    const SESSION_DURATION = 15 * 60;
    let remainingSeconds = SESSION_DURATION;

    let _sseConnection = null;
    let _refreshDebounceTimer = null;
    let timerInterval = null;
    let reportPollingInterval = null;
    let notifPollingInterval = null;

    let offlineBanner = null;
    let offlineRestoreTimer = null;
    let hostRecoveryTimer = null;
    let isCheckingHost = false;
    let isHostReachable = true;

    function showOfflineBanner() {
        isHostReachable = false;
        if (offlineRestoreTimer) {
            clearTimeout(offlineRestoreTimer);
            offlineRestoreTimer = null;
        }
        if (offlineBanner) {
            if (!offlineBanner.classList.contains('online-restored')) return;
            offlineBanner.className = 'offline-banner';
        } else {
            offlineBanner = document.createElement('div');
            offlineBanner.className = 'offline-banner';
            document.body.prepend(offlineBanner);
        }
        offlineBanner.innerHTML = `
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m2 2 20 20"/><path d="M8.5 16.5a5 5 0 0 1 7 0"/><path d="M2 8.82a15 15 0 0 1 4.17-2.65"/><path d="M10.66 5c4.01-.36 8.14.9 11.34 3.76"/><path d="M16.85 11.25a10 10 0 0 1 2.22 1.68"/><path d="M5 13a10 10 0 0 1 5.24-2.78"/><line x1="12" y1="20" x2="12.01" y2="20"/></svg>
            <span>ارتباط با سرور قطع شده است. لطفاً اتصال اینترنت خود را بررسی کنید.</span>
        `;
        forceLogoutOnConnectionDrop();
        startRecoveryPolling();
    }

    function hideOfflineBanner() {
        if (!offlineBanner) return;
        isHostReachable = true;
        stopRecoveryPolling();
        offlineBanner.className = 'offline-banner online-restored';
        offlineBanner.innerHTML = `
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h.01"/><path d="M2 8.82a15 15 0 0 1 20 0"/><path d="M5 12.86a10 10 0 0 1 14 0"/><path d="M8.5 16.5a5 5 0 0 1 7 0"/></svg>
            <span>ارتباط با سرور برقرار شد.</span>
        `;
        if (typeof loadAdminCaptcha === 'function') {
            loadAdminCaptcha();
        }
        if (offlineRestoreTimer) clearTimeout(offlineRestoreTimer);
        offlineRestoreTimer = setTimeout(() => {
            if (offlineBanner && offlineBanner.classList.contains('online-restored')) {
                offlineBanner.remove();
                offlineBanner = null;
            }
            offlineRestoreTimer = null;
        }, 3000);
    }

    async function pingHost(timeoutMs = 4000) {
        const ctrl = new AbortController();
        const timer = setTimeout(() => ctrl.abort(), timeoutMs);
        try {
            const res = await fetch('/health', {
                method: 'GET',
                cache: 'no-store',
                signal: ctrl.signal
            });
            return res.status >= 200 && res.status < 500;
        } catch {
            return false;
        } finally {
            clearTimeout(timer);
        }
    }

    async function checkHostConnection() {
        if (isCheckingHost) return isHostReachable;
        if (!navigator.onLine) {
            showOfflineBanner();
            return false;
        }
        isCheckingHost = true;
        try {
            const reachable = await pingHost(4000);
            if (reachable) {
                if (offlineBanner && !offlineBanner.classList.contains('online-restored')) {
                    hideOfflineBanner();
                }
            } else {
                showOfflineBanner();
            }
            return reachable;
        } finally {
            isCheckingHost = false;
        }
    }

    function startRecoveryPolling() {
        if (hostRecoveryTimer) return;
        hostRecoveryTimer = setInterval(() => {
            if (document.visibilityState === 'visible') {
                checkHostConnection();
            }
        }, 5000);
    }

    function stopRecoveryPolling() {
        if (hostRecoveryTimer) {
            clearInterval(hostRecoveryTimer);
            hostRecoveryTimer = null;
        }
    }

    function forceLogoutOnConnectionDrop() {
        const isDashboardActive = dashboardSection && dashboardSection.style.display !== 'none' && !dashboardSection.classList.contains('d-none');
        if (isDashboardActive || currentAdmin !== null) {
            try {
                fetch('/api/admin/logout', { method: 'POST', credentials: 'same-origin', keepalive: true }).catch(() => {});
            } catch (_) {}
            document.cookie = 'herasat_admin_session=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
            sessionStorage.removeItem('herasat_csrf');
            sessionStorage.removeItem('2fa_prompt_shown');
            currentAdmin = null;
            if (_sseConnection) {
                _sseConnection.close();
                _sseConnection = null;
            }
            if (timerInterval) {
                clearInterval(timerInterval);
                timerInterval = null;
            }
            if (typeof stopBackgroundPolling === 'function') {
                stopBackgroundPolling();
            }
            if (reportPollingInterval) {
                clearInterval(reportPollingInterval);
                reportPollingInterval = null;
            }
            if (notifPollingInterval) {
                clearInterval(notifPollingInterval);
                notifPollingInterval = null;
            }
            document.querySelectorAll('.admin-modal-overlay').forEach(m => m.style.display = 'none');
            if (typeof cancelMFAStep === 'function') {
                cancelMFAStep();
            }
            if (dashboardSection) {
                dashboardSection.style.display = 'none';
                dashboardSection.classList.add('d-none');
            }
            if (loginSection) {
                loginSection.style.display = 'flex';
                loginSection.classList.remove('d-none');
            }
            if (loginMessage) {
                loginMessage.textContent = 'به دلیل قطع ارتباط با سرور و جهت حفظ امنیت، نشست کاری شما پایان یافت. لطفاً پس از اتصال مجدد، دوباره وارد شوید.';
                loginMessage.style.display = 'block';
            }
            if (typeof loadAdminCaptcha === 'function') {
                loadAdminCaptcha();
            }
        } else {
            if (typeof cancelMFAStep === 'function') {
                const mfaStep = document.getElementById('mfaStepContainer');
                if (mfaStep && mfaStep.style.display !== 'none') {
                    cancelMFAStep();
                }
            }
        }
    }

    window.__checkAdminHostConnection = checkHostConnection;

    window.addEventListener('offline', () => {
        showOfflineBanner();
    });

    window.addEventListener('online', () => {
        checkHostConnection();
    });

    setInterval(() => {
        if (document.visibilityState === 'visible' && !offlineBanner) {
            checkHostConnection();
        }
    }, 45000);

    if (!navigator.onLine) {
        showOfflineBanner();
    }

    const countdownEl = document.getElementById('sessionCountdown');
    const timerBoxEl = document.getElementById('sessionTimerBox') || document.querySelector('.session-timer-box');
    const btnRenewSession = document.getElementById('btnRenewSession');

    const sessionWarningModal = document.getElementById('sessionWarningModal');
    const warningCountdownText = document.getElementById('warningCountdownText');
    const btnExtendSessionWarning = document.getElementById('btnExtendSessionWarning');
    const btnLogoutSessionWarning = document.getElementById('btnLogoutSessionWarning');

    function updateTimerDisplay() {
        if (!countdownEl) return;
        const minutes = Math.floor(remainingSeconds / 60);
        const seconds = remainingSeconds % 60;
        countdownEl.textContent = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;

        const statusDot = document.querySelector('.status-indicator-dot');
        const statusText = document.getElementById('sessionStatusText');

        if (remainingSeconds <= 120) {
            timerBoxEl?.classList.add('warning');
            if (statusDot) statusDot.className = 'status-indicator-dot danger';
            if (statusText) statusText.textContent = 'وضعیت: در شرف انقضا!';
        } else if (remainingSeconds <= 300) {
            timerBoxEl?.classList.remove('warning');
            if (statusDot) statusDot.className = 'status-indicator-dot warning';
            if (statusText) statusText.textContent = 'وضعیت: نیاز به تمدید';
        } else {
            timerBoxEl?.classList.remove('warning');
            if (statusDot) statusDot.className = 'status-indicator-dot';
            if (statusText) statusText.textContent = 'وضعیت: فعال و امن';
        }
    }

    function resetSessionTimer() {
        remainingSeconds = SESSION_DURATION;
        if (sessionWarningModal) sessionWarningModal.style.display = 'none';
        updateTimerDisplay();
    }

    async function renewSessionInteractive(e) {
        if (e) e.stopPropagation();
        const icon = btnRenewSession?.querySelector('svg');
        if (icon) icon.classList.add('spinning');

        try {
            await fetchReports();
            if (sessionWarningModal) sessionWarningModal.style.display = 'none';
            showToast('اعتبار نشست با موفقیت به ۱۵ دقیقه تمدید شد', 2500, 'success');
        } catch (err) {
            showToast('خطا در تمدید نشست. ارتباط با سرور برقرار نشد.', 3500, 'error');
        } finally {
            setTimeout(() => {
                if (icon) icon.classList.remove('spinning');
            }, 500);
        }
    }

    if (btnRenewSession) {
        btnRenewSession.addEventListener('click', renewSessionInteractive);
    }
    if (btnExtendSessionWarning) {
        btnExtendSessionWarning.addEventListener('click', renewSessionInteractive);
    }
    if (btnLogoutSessionWarning) {
        btnLogoutSessionWarning.addEventListener('click', async () => {
            try {
                await adminFetch('/api/admin/logout', { method: 'POST' });
            } catch (e) { }
            document.cookie = 'herasat_admin_session=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
            location.reload();
        });
    }

    if (timerBoxEl) {
        timerBoxEl.addEventListener('click', (e) => {
            if (e.target.closest('#btnRenewSession')) return;
            renewSessionInteractive(e);
        });
        timerBoxEl.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                renewSessionInteractive(e);
            }
        });
    }

    function handleTimerTick() {
        remainingSeconds--;
        updateTimerDisplay();

        if (remainingSeconds <= 60 && remainingSeconds > 0) {
            if (sessionWarningModal && sessionWarningModal.style.display !== 'flex') {
                sessionWarningModal.style.display = 'flex';
            }
            if (warningCountdownText) warningCountdownText.textContent = `${toPersianNum(remainingSeconds)} ثانیه`;
        } else if (remainingSeconds > 60) {
            if (sessionWarningModal) sessionWarningModal.style.display = 'none';
        }

        if (remainingSeconds <= 0) {
            clearInterval(timerInterval);
            timerInterval = null;
            showConfirmDialog({
                title: 'پایان نشست کاری',
                description: 'نشست کاری شما به دلیل عدم فعالیت به پایان رسید. لطفاً مجدداً وارد شوید.',
                confirmText: 'ورود مجدد به سامانه',
                cancelText: 'بستن',
                variant: 'warning',
                iconType: 'shield-alert'
            }).then(() => {
                forceLogoutOnConnectionDrop();
            });
        }
    }

    let lastUserActivityReset = Date.now();
    function onUserInteraction() {
        const now = Date.now();
        if (now - lastUserActivityReset > 25000) {
            lastUserActivityReset = now;
            if (dashboardSection && dashboardSection.style.display !== 'none' && !dashboardSection.classList.contains('d-none') && currentAdmin) {
                resetSessionTimer();
            }
        }
    }
    ['mousemove', 'keydown', 'click', 'scroll', 'touchstart'].forEach(evt => {
        window.addEventListener(evt, onUserInteraction, { passive: true });
    });

    function resumeSessionTimer() {
        clearInterval(timerInterval);
        if (remainingSeconds > 0) {
            timerInterval = setInterval(handleTimerTick, 1000);
        }
    }

    function startSessionTimer() {
        clearInterval(timerInterval);
        resetSessionTimer();
        timerInterval = setInterval(handleTimerTick, 1000);
    }

    const adminCaptchaImg = document.getElementById('admin-captcha-img');
    const adminCaptchaId = document.getElementById('admin_captcha_id');
    const adminCaptchaAnswer = document.getElementById('admin_captcha_answer');
    const adminCaptchaRefresh = document.getElementById('admin-captcha-refresh');

    async function loadAdminCaptcha() {
        if (adminCaptchaImg) adminCaptchaImg.classList.add('skeleton-shimmer');
        try {
            const res = await adminFetch('/api/captcha');
            if (res.ok) {
                const data = await res.json();
                if (adminCaptchaImg) {
                    adminCaptchaImg.src = data.image_base64;
                    adminCaptchaImg.onload = () => adminCaptchaImg.classList.remove('skeleton-shimmer');
                }
                if (adminCaptchaId) adminCaptchaId.value = data.captcha_id;
                if (adminCaptchaAnswer) adminCaptchaAnswer.value = '';
            } else {
                if (adminCaptchaImg) adminCaptchaImg.classList.remove('skeleton-shimmer');
            }
        } catch (err) {
            console.error('Failed to load CAPTCHA');
            if (adminCaptchaImg) adminCaptchaImg.classList.remove('skeleton-shimmer');
        }
    }

    if (adminCaptchaRefresh) {
        adminCaptchaRefresh.addEventListener('click', loadAdminCaptcha);
    }
    loadAdminCaptcha();

    const togglePasswordBtn = document.getElementById('togglePasswordBtn');
    const passwordInput = document.getElementById('password');
    const eyeShow = document.getElementById('eye-icon-show');
    const eyeHide = document.getElementById('eye-icon-hide');
    if (togglePasswordBtn && passwordInput) {
        togglePasswordBtn.addEventListener('click', () => {
            const isPassword = passwordInput.type === 'password';
            passwordInput.type = isPassword ? 'text' : 'password';
            if (eyeShow && eyeHide) {
                eyeShow.classList.toggle('d-none', isPassword);
                eyeHide.classList.toggle('d-none', !isPassword);
            }
        });
    }

    function initLoginSlider() {
        const sliderCol = document.getElementById('loginSliderCol');
        if (!sliderCol) return;
        const slides = sliderCol.querySelectorAll('.admin-slider-slide');
        const dots = sliderCol.querySelectorAll('.admin-slider-dot');
        if (!slides.length) return;

        let currentIndex = 0;
        let slideTimer = null;

        function goToSlide(index) {
            if (index < 0 || index >= slides.length) return;
            slides[currentIndex]?.classList.remove('active');
            dots[currentIndex]?.classList.remove('active');
            currentIndex = index;
            slides[currentIndex]?.classList.add('active');
            dots[currentIndex]?.classList.add('active');
        }

        function nextSlide() {
            const nextIdx = (currentIndex + 1) % slides.length;
            goToSlide(nextIdx);
        }

        function startTimer() {
            stopTimer();
            slideTimer = setInterval(() => {
                if (document.hidden) return;
                if (!loginSection || loginSection.style.display === 'none' || loginSection.classList.contains('d-none')) {
                    stopTimer();
                    return;
                }
                nextSlide();
            }, 4500);
        }

        function stopTimer() {
            if (slideTimer) {
                clearInterval(slideTimer);
                slideTimer = null;
            }
        }

        dots.forEach((dot, idx) => {
            dot.addEventListener('click', () => {
                goToSlide(idx);
                startTimer();
            });
        });

        sliderCol.addEventListener('mouseenter', stopTimer, { passive: true });
        sliderCol.addEventListener('mouseleave', () => {
            if (loginSection && loginSection.style.display !== 'none' && !loginSection.classList.contains('d-none')) {
                startTimer();
            }
        }, { passive: true });

        startTimer();
    }
    initLoginSlider();

    (async () => {
        try {
            const hasSession = await checkAdminProfile();
            if (hasSession) {
                if (loginSection) {
                    loginSection.classList.add('d-none');
                    loginSection.style.display = 'none';
                }
                if (dashboardSection) {
                    dashboardSection.classList.remove('d-none');
                    dashboardSection.style.display = 'block';
                }
                startSessionTimer();
                startBackgroundPolling();
                switchMainTab('reports');
                runWhenIdle(fetchNotifications);
            }
        } catch (e) {
            console.error(e);
        } finally {
            window.__adminStartupComplete = true;
            if (document.readyState === 'complete') {
                if (typeof window.markAdminReady === 'function') window.markAdminReady();
            } else {
                window.addEventListener('load', () => {
                    if (typeof window.markAdminReady === 'function') window.markAdminReady();
                });
            }
        }
    })();

    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        if (!navigator.onLine || !isHostReachable) {
            showOfflineBanner();
            if (loginMessage) {
                loginMessage.textContent = 'ارتباط با سرور قطع شده است. لطفاً اتصال اینترنت خود را بررسی کنید.';
                loginMessage.style.display = 'block';
            }
            return;
        }

        const loginBtn = document.getElementById('loginBtn');
        const btnText = loginBtn?.querySelector('.btn-text');
        const btnSpinner = loginBtn?.querySelector('.btn-spinner');

        if (loginBtn) loginBtn.disabled = true;
        if (btnText) btnText.textContent = 'در حال احراز هویت...';
        if (btnSpinner) btnSpinner.classList.remove('d-none');
        if (loginMessage) loginMessage.style.display = 'none';

        const payload = {
            username: loginForm.username.value,
            password: loginForm.password.value,
            captcha_id: adminCaptchaId.value,
            captcha_answer: adminCaptchaAnswer.value
        };

        try {
            const response = await adminFetch('/api/admin/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (response.ok) {
                const data = await response.clone().json().catch(() => null);
                if (data && data.status === 'mfa_required') {
                    showMFAStep(data.mfa_token);
                    return;
                }
                if (data && data.csrf_token) {
                    sessionStorage.setItem('herasat_csrf', data.csrf_token);
                }
                sessionStorage.removeItem('2fa_prompt_shown');
                if (loginSection) {
                    loginSection.classList.add('d-none');
                    loginSection.style.display = 'none';
                }
                if (dashboardSection) {
                    dashboardSection.classList.remove('d-none');
                    dashboardSection.style.display = 'block';
                }
                startSessionTimer();
                startBackgroundPolling();
                await checkAdminProfile();
                switchMainTab('reports');
                runWhenIdle(fetchNotifications);
            } else {
                const errText = await response.text();
                loginMessage.textContent = errText || 'نام کاربری یا رمز عبور اشتباه است.';
                loginMessage.style.display = 'block';
                loadAdminCaptcha();
            }
        } catch (error) {
            loginMessage.textContent = 'خطای ارتباط با سرور.';
            loginMessage.style.display = 'block';
            loadAdminCaptcha();
        } finally {
            if (loginBtn) loginBtn.disabled = false;
            if (btnText) btnText.textContent = 'ورود به سیستم';
            if (btnSpinner) btnSpinner.classList.add('d-none');
        }
    });

    let mfaCountdownInterval = null;
    let mfaOtpController = null;

    function createOtpController(container, hiddenInput, formEl, autoSubmit = true) {
        if (!container || !hiddenInput) return null;

        const cells = Array.from(container.querySelectorAll('.otp-cell'));

        function syncValue() {
            const val = cells.map(c => toAsciiDigits(c.value.trim())).join('');
            hiddenInput.value = val;
            cells.forEach(c => c.classList.toggle('has-value', c.value.trim().length > 0));
            return val;
        }

        function clearAll() {
            cells.forEach(c => {
                c.value = '';
                c.classList.remove('has-value');
            });
            hiddenInput.value = '';
            container.classList.remove('is-error', 'is-success');
        }

        function showError() {
            container.classList.remove('is-success');
            container.classList.add('is-error');
            setTimeout(() => {
                container.classList.remove('is-error');
                clearAll();
                if (cells[0]) {
                    cells[0].focus();
                    cells[0].select();
                }
            }, 600);
        }

        function showSuccess() {
            container.classList.remove('is-error');
            container.classList.add('is-success');
        }

        cells.forEach((cell, idx) => {
            cell.addEventListener('input', () => {
                container.classList.remove('is-error');
                const raw = toAsciiDigits(cell.value);
                const digits = raw.replace(/[^0-9]/g, '');

                if (digits.length === 0) {
                    cell.value = '';
                    syncValue();
                    return;
                }

                if (digits.length === 1) {
                    cell.value = digits;
                    const val = syncValue();
                    if (idx < cells.length - 1) {
                        cells[idx + 1].focus();
                        cells[idx + 1].select();
                    } else if (val.length === 6 && autoSubmit && formEl) {
                        formEl.requestSubmit ? formEl.requestSubmit() : formEl.dispatchEvent(new Event('submit', { cancelable: true }));
                    }
                    return;
                }

                let cursor = idx;
                for (let i = 0; i < digits.length && cursor < cells.length; i++) {
                    cells[cursor].value = digits[i];
                    cursor++;
                }
                const val = syncValue();
                if (cursor < cells.length) {
                    cells[cursor].focus();
                    cells[cursor].select();
                } else {
                    cells[cells.length - 1].focus();
                    cells[cells.length - 1].select();
                    if (val.length === 6 && autoSubmit && formEl) {
                        formEl.requestSubmit ? formEl.requestSubmit() : formEl.dispatchEvent(new Event('submit', { cancelable: true }));
                    }
                }
            });

            cell.addEventListener('keydown', (e) => {
                container.classList.remove('is-error');
                if (e.key === 'Backspace') {
                    if (cell.value) {
                        cell.value = '';
                        syncValue();
                        e.preventDefault();
                    } else if (idx > 0) {
                        cells[idx - 1].value = '';
                        cells[idx - 1].focus();
                        cells[idx - 1].select();
                        syncValue();
                        e.preventDefault();
                    }
                } else if (e.key === 'Delete') {
                    cell.value = '';
                    syncValue();
                    e.preventDefault();
                } else if (e.key === 'ArrowLeft') {
                    e.preventDefault();
                    if (idx > 0) {
                        cells[idx - 1].focus();
                        cells[idx - 1].select();
                    }
                } else if (e.key === 'ArrowRight') {
                    e.preventDefault();
                    if (idx < cells.length - 1) {
                        cells[idx + 1].focus();
                        cells[idx + 1].select();
                    }
                } else if (e.key === 'Home') {
                    e.preventDefault();
                    cells[0].focus();
                    cells[0].select();
                } else if (e.key === 'End') {
                    e.preventDefault();
                    cells[cells.length - 1].focus();
                    cells[cells.length - 1].select();
                } else if (e.key === 'Enter') {
                    if (formEl) {
                        e.preventDefault();
                        formEl.requestSubmit ? formEl.requestSubmit() : formEl.dispatchEvent(new Event('submit', { cancelable: true }));
                    }
                }
            });

            cell.addEventListener('paste', (e) => {
                e.preventDefault();
                container.classList.remove('is-error');
                const text = (e.clipboardData || window.clipboardData).getData('text');
                const digits = toAsciiDigits(text).replace(/[^0-9]/g, '').slice(0, 6);
                if (!digits) return;

                const startIdx = digits.length === 6 ? 0 : idx;
                let cursor = startIdx;
                for (let i = 0; i < digits.length && cursor < cells.length; i++) {
                    cells[cursor].value = digits[i];
                    cursor++;
                }
                const val = syncValue();
                if (cursor < cells.length) {
                    cells[cursor].focus();
                    cells[cursor].select();
                } else {
                    cells[cells.length - 1].focus();
                    cells[cells.length - 1].select();
                    if (val.length === 6 && autoSubmit && formEl) {
                        formEl.requestSubmit ? formEl.requestSubmit() : formEl.dispatchEvent(new Event('submit', { cancelable: true }));
                    }
                }
            });

            cell.addEventListener('focus', () => {
                cell.select();
            });
        });

        return {
            clear: clearAll,
            showError,
            showSuccess,
            focus: () => {
                const firstEmpty = cells.find(c => !c.value.trim());
                if (firstEmpty) {
                    firstEmpty.focus();
                    firstEmpty.select();
                } else {
                    cells[0].focus();
                    cells[0].select();
                }
            }
        };
    }

    function initMfaOtpInput() {
        return createOtpController(
            document.getElementById('mfaOtpContainer'),
            document.getElementById('mfa_totp_code'),
            document.getElementById('mfaForm'),
            true
        );
    }

    function showMFAStep(token) {
        const mfaStep = document.getElementById('mfaStepContainer');
        const mfaPendingToken = document.getElementById('mfa_pending_token');
        const mfaMsg = document.getElementById('mfaMessage');
        const countdownEl = document.getElementById('mfaCountdownText');

        if (loginForm) loginForm.style.display = 'none';
        if (loginMessage) loginMessage.style.display = 'none';
        if (mfaStep) mfaStep.style.display = 'block';
        if (mfaPendingToken) mfaPendingToken.value = token;

        if (!mfaOtpController) {
            mfaOtpController = initMfaOtpInput();
        }
        if (mfaOtpController) {
            mfaOtpController.clear();
            setTimeout(() => mfaOtpController.focus(), 80);
        }

        if (mfaMsg) mfaMsg.style.display = 'none';

        let secondsRemaining = 180;
        if (mfaCountdownInterval) clearInterval(mfaCountdownInterval);
        mfaCountdownInterval = setInterval(() => {
            secondsRemaining--;
            if (secondsRemaining <= 0) {
                clearInterval(mfaCountdownInterval);
                if (mfaMsg) {
                    mfaMsg.textContent = 'زمان اعتبار کد منقضی شد. لطفاً مجدداً وارد شوید.';
                    mfaMsg.className = 'response-box error';
                    mfaMsg.style.display = 'block';
                }
                setTimeout(cancelMFAStep, 2000);
            } else if (countdownEl) {
                const m = Math.floor(secondsRemaining / 60);
                const s = secondsRemaining % 60;
                countdownEl.textContent = `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
            }
        }, 1000);
    }

    function cancelMFAStep() {
        if (mfaCountdownInterval) clearInterval(mfaCountdownInterval);
        const mfaStep = document.getElementById('mfaStepContainer');
        if (mfaStep) mfaStep.style.display = 'none';
        if (loginForm) loginForm.style.display = 'block';
        if (mfaOtpController) mfaOtpController.clear();
        loadAdminCaptcha();
    }

    const btnCancelMFA = document.getElementById('btnCancelMFA');
    if (btnCancelMFA) {
        btnCancelMFA.addEventListener('click', cancelMFAStep);
    }

    const mfaForm = document.getElementById('mfaForm');
    if (mfaForm) {
        mfaForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            if (!navigator.onLine || !isHostReachable) {
                showOfflineBanner();
                const mfaMsg = document.getElementById('mfaMessage');
                if (mfaMsg) {
                    mfaMsg.textContent = 'ارتباط با سرور قطع شده است. لطفاً اتصال اینترنت خود را بررسی کنید.';
                    mfaMsg.className = 'response-box error';
                    mfaMsg.style.display = 'block';
                }
                return;
            }

            const mfaPendingToken = document.getElementById('mfa_pending_token');
            const mfaTotpCode = document.getElementById('mfa_totp_code');
            const mfaMsg = document.getElementById('mfaMessage');
            const mfaSubmitBtn = document.getElementById('mfaSubmitBtn');

            const token = mfaPendingToken ? mfaPendingToken.value : '';
            const code = mfaTotpCode ? mfaTotpCode.value.trim() : '';

            if (code.length !== 6) {
                if (mfaMsg) {
                    mfaMsg.textContent = 'لطفاً کد ۶ رقمی امنیتی را به صورت کامل وارد نمایید.';
                    mfaMsg.className = 'response-box error';
                    mfaMsg.style.display = 'block';
                }
                if (mfaOtpController) mfaOtpController.showError();
                return;
            }

            if (mfaSubmitBtn) mfaSubmitBtn.disabled = true;
            if (mfaMsg) mfaMsg.style.display = 'none';

            try {
                const res = await adminFetch('/api/admin/login/verify-mfa', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ mfa_token: token, totp_code: code })
                });

                if (res.ok) {
                    if (mfaOtpController) mfaOtpController.showSuccess();
                    const data = await res.json();
                    if (data && data.csrf_token) {
                        sessionStorage.setItem('herasat_csrf', data.csrf_token);
                    }
                    if (mfaCountdownInterval) clearInterval(mfaCountdownInterval);
                    const mfaStep = document.getElementById('mfaStepContainer');
                    if (mfaStep) mfaStep.style.display = 'none';
                    if (loginSection) {
                        loginSection.classList.add('d-none');
                        loginSection.style.display = 'none';
                    }
                    if (dashboardSection) {
                        dashboardSection.classList.remove('d-none');
                        dashboardSection.style.display = 'block';
                    }
                    startSessionTimer();
                    startBackgroundPolling();
                    await checkAdminProfile();
                    switchMainTab('reports');
                    runWhenIdle(fetchNotifications);
                } else {
                    const errText = await res.text();
                    if (mfaOtpController) mfaOtpController.showError();
                    if (mfaMsg) {
                        mfaMsg.textContent = errText || 'کد امنیتی اشتباه است یا منقضی شده است.';
                        mfaMsg.className = 'response-box error';
                        mfaMsg.style.display = 'block';
                    }
                }
            } catch (err) {
                if (mfaOtpController) mfaOtpController.showError();
                if (mfaMsg) {
                    mfaMsg.textContent = 'خطای ارتباط با سرور.';
                    mfaMsg.className = 'response-box error';
                    mfaMsg.style.display = 'block';
                }
            } finally {
                if (mfaSubmitBtn) mfaSubmitBtn.disabled = false;
            }
        });
    }

    const btnOpen2FAModal = document.getElementById('btnOpen2FAModal');
    const modal2FA = document.getElementById('modal-2fa-management');
    const btnClose2FAModal = document.getElementById('btn-close-2fa-modal');
    const btnStartSetup2FA = document.getElementById('btn-start-setup-2fa');
    const btnStartDisable2FA = document.getElementById('btn-start-disable-2fa');
    const setupSection2FA = document.getElementById('2fa-setup-section');
    const disableSection2FA = document.getElementById('2fa-disable-section');
    const initialActions2FA = document.getElementById('2fa-initial-actions');
    const formEnable2FA = document.getElementById('form-enable-2fa');
    const formDisable2FA = document.getElementById('form-disable-2fa');
    const btnCancelSetup2FA = document.getElementById('btn-cancel-setup-2fa');
    const btnCancelDisable2FA = document.getElementById('btn-cancel-disable-2fa');

    let currentAdmin2FAEnabled = false;
    let enable2faOtpController = null;
    let disable2faOtpController = null;

    function init2FAModalOtpControllers() {
        if (!enable2faOtpController) {
            enable2faOtpController = createOtpController(
                document.getElementById('enable2faOtpContainer'),
                document.getElementById('enable_2fa_code'),
                formEnable2FA,
                false
            );
        }
        if (!disable2faOtpController) {
            disable2faOtpController = createOtpController(
                document.getElementById('disable2faOtpContainer'),
                document.getElementById('disable_2fa_code'),
                formDisable2FA,
                false
            );
        }
    }

    if (btnOpen2FAModal) {
        btnOpen2FAModal.addEventListener('click', async () => {
            init2FAModalOtpControllers();
            if (modal2FA) modal2FA.style.display = 'flex';
            reset2FAModalViews();
            await refresh2FAStatus();
        });
    }

    function reset2FAModalViews() {
        if (setupSection2FA) setupSection2FA.style.display = 'none';
        if (disableSection2FA) disableSection2FA.style.display = 'none';
        if (initialActions2FA) initialActions2FA.style.display = 'flex';
        const enableMsg = document.getElementById('enable-2fa-status');
        if (enableMsg) enableMsg.style.display = 'none';
        const disableMsg = document.getElementById('disable-2fa-status');
        if (disableMsg) disableMsg.style.display = 'none';
        const epwd = document.getElementById('enable_2fa_pwd');
        if (epwd) epwd.value = '';
        const ecode = document.getElementById('enable_2fa_code');
        if (ecode) ecode.value = '';
        const dpwd = document.getElementById('disable_2fa_pwd');
        if (dpwd) dpwd.value = '';
        const dcode = document.getElementById('disable_2fa_code');
        if (dcode) dcode.value = '';
        if (enable2faOtpController) enable2faOtpController.clear();
        if (disable2faOtpController) disable2faOtpController.clear();
        if (typeof resetPasswordWrappers === 'function') resetPasswordWrappers(modal2FA);
    }

    async function refresh2FAStatus() {
        try {
            const res = await adminFetch('/api/admin/me');
            if (res.ok) {
                const data = await res.json();
                currentAdmin2FAEnabled = !!data.totp_enabled;
                const activeBanner = document.getElementById('2fa-active-banner');
                const inactiveBanner = document.getElementById('2fa-inactive-banner');
                if (currentAdmin2FAEnabled) {
                    if (activeBanner) activeBanner.style.display = 'block';
                    if (inactiveBanner) inactiveBanner.style.display = 'none';
                    if (btnStartSetup2FA) btnStartSetup2FA.style.display = 'none';
                    if (btnStartDisable2FA) btnStartDisable2FA.style.display = 'inline-flex';
                } else {
                    if (activeBanner) activeBanner.style.display = 'none';
                    if (inactiveBanner) inactiveBanner.style.display = 'block';
                    if (btnStartSetup2FA) btnStartSetup2FA.style.display = 'inline-flex';
                    if (btnStartDisable2FA) btnStartDisable2FA.style.display = 'none';
                }
            }
        } catch (_) { }
    }

    if (btnClose2FAModal) {
        btnClose2FAModal.addEventListener('click', () => {
            if (modal2FA) modal2FA.style.display = 'none';
        });
    }
    if (btnCancelSetup2FA) {
        btnCancelSetup2FA.addEventListener('click', reset2FAModalViews);
    }
    if (btnCancelDisable2FA) {
        btnCancelDisable2FA.addEventListener('click', reset2FAModalViews);
    }

    if (btnStartSetup2FA) {
        btnStartSetup2FA.addEventListener('click', async () => {
            try {
                const res = await adminFetch('/api/admin/2fa/setup', { method: 'POST' });
                if (res.ok) {
                    const data = await res.json();
                    const secretDisplay = document.getElementById('2fa-secret-display');
                    const hiddenSecret = document.getElementById('enable_2fa_secret');
                    if (secretDisplay && data.secret) {
                        const spaced = data.secret.match(/.{1,4}/g)?.join(' ') || data.secret;
                        secretDisplay.textContent = spaced;
                    }
                    if (hiddenSecret) hiddenSecret.value = data.secret;
                    init2FAModalOtpControllers();
                    if (enable2faOtpController) enable2faOtpController.clear();
                    if (setupSection2FA) setupSection2FA.style.display = 'block';
                    if (initialActions2FA) initialActions2FA.style.display = 'none';
                    setTimeout(() => {
                        const pwdInput = document.getElementById('enable_2fa_pwd');
                        if (pwdInput && !pwdInput.value) {
                            pwdInput.focus();
                        } else if (enable2faOtpController) {
                            enable2faOtpController.focus();
                        }
                    }, 120);
                } else {
                    showToast('خطا در ایجاد کلید احراز هویت دومرحله‌ای', 3000, 'error');
                }
            } catch (_) {
                showToast('خطای ارتباط با سرور', 3000, 'error');
            }
        });
    }

    function toAsciiDigits(str) {
        if (!str) return '';
        return String(str)
            .replace(/[۰-۹]/g, d => String.fromCharCode(d.charCodeAt(0) - 1728))
            .replace(/[٠-٩]/g, d => String.fromCharCode(d.charCodeAt(0) - 1584))
            .replace(/[^0-9]/g, '');
    }

    const mfaCodeInput = document.getElementById('enable_2fa_code');
    if (mfaCodeInput) {
        mfaCodeInput.addEventListener('input', () => {
            mfaCodeInput.value = toAsciiDigits(mfaCodeInput.value);
        });
    }

    const mfaLoginCodeInput = document.getElementById('mfa_totp_code');
    if (mfaLoginCodeInput) {
        mfaLoginCodeInput.addEventListener('input', () => {
            mfaLoginCodeInput.value = toAsciiDigits(mfaLoginCodeInput.value);
        });
    }

    const btnCopy2FASecret = document.getElementById('btn-copy-2fa-secret');
    if (btnCopy2FASecret) {
        btnCopy2FASecret.addEventListener('click', () => {
            const secret = document.getElementById('enable_2fa_secret')?.value || document.getElementById('2fa-secret-display')?.textContent;
            if (secret && secret !== '...') {
                navigator.clipboard.writeText(secret).then(() => {
                    const textSpan = document.getElementById('copy-secret-text');
                    if (textSpan) {
                        const original = textSpan.textContent;
                        textSpan.textContent = 'کپی شد!';
                        setTimeout(() => { textSpan.textContent = original; }, 2000);
                    }
                    showToast('کلید امنیتی با موفقیت کپی شد', 2000, 'success');
                }).catch(() => {
                    showToast('امکان کپی خودکار فراهم نشد. لطفاً متن را دستی کپی کنید.', 3000, 'warning');
                });
            }
        });
    }

    function triggerPasswordError(pwdInput) {
        if (!pwdInput) return;
        const wrapper = pwdInput.closest('.password-wrapper') || pwdInput;
        pwdInput.classList.add('input-error');
        wrapper.classList.add('input-error', 'input-error-shake');
        setTimeout(() => wrapper.classList.remove('input-error-shake'), 450);
        pwdInput.focus();
        if (typeof pwdInput.select === 'function') pwdInput.select();
    }

    ['enable_2fa_pwd', 'disable_2fa_pwd'].forEach(id => {
        const el = document.getElementById(id);
        if (el) {
            el.addEventListener('input', () => {
                el.classList.remove('input-error');
                el.closest('.password-wrapper')?.classList.remove('input-error');
                const statusEl = el.closest('form')?.querySelector('.response-box');
                if (statusEl) statusEl.style.display = 'none';
            });
        }
    });

    if (formEnable2FA) {
        formEnable2FA.addEventListener('submit', async (e) => {
            e.preventDefault();
            const secret = document.getElementById('enable_2fa_secret').value;
            const rawCode = document.getElementById('enable_2fa_code').value;
            const code = toAsciiDigits(rawCode);
            const pwdInput = document.getElementById('enable_2fa_pwd');
            const pwd = pwdInput ? pwdInput.value : '';
            const statusEl = document.getElementById('enable-2fa-status');

            if (!pwd || !pwd.trim()) {
                triggerPasswordError(pwdInput);
                if (statusEl) {
                    statusEl.textContent = 'لطفاً کلمه عبور جاری حساب کاربری را وارد نمایید.';
                    statusEl.className = 'response-box error';
                    statusEl.style.display = 'block';
                }
                return;
            }

            if (!code || code.length < 6) {
                if (enable2faOtpController) enable2faOtpController.showError();
                if (statusEl) {
                    statusEl.textContent = 'لطفاً کد ۶ رقمی نرم‌افزار را به طور کامل وارد نمایید.';
                    statusEl.className = 'response-box error';
                    statusEl.style.display = 'block';
                }
                return;
            }

            try {
                const res = await adminFetch('/api/admin/2fa/enable', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ password: pwd, secret, totp_code: code })
                });
                if (res.ok) {
                    if (enable2faOtpController) enable2faOtpController.showSuccess();
                    if (currentAdmin) currentAdmin.totp_enabled = true;
                    showToast('احراز هویت دومرحله‌ای با موفقیت فعال شد', 4000, 'success');
                    reset2FAModalViews();
                    await refresh2FAStatus();
                } else {
                    const err = (await res.text()).trim();
                    const isPwdErr = err.includes('کلمه عبور') || err.includes('رمز عبور') || (res.status === 401 && !err.includes('منقضی') && !err.includes('Unauthorized'));
                    if (isPwdErr) {
                        triggerPasswordError(pwdInput);
                    } else {
                        if (enable2faOtpController) enable2faOtpController.showError();
                    }
                    if (statusEl) {
                        if (err === 'Unauthorized' || (res.status === 401 && !isPwdErr)) {
                            statusEl.textContent = 'نشست کاری شما منقضی شده است. در حال انتقال به صفحه ورود...';
                            setTimeout(() => location.reload(), 1500);
                        } else {
                            statusEl.textContent = err || (isPwdErr ? 'کلمه عبور جاری اشتباه است.' : 'کد امنیتی تأیید نشد.');
                        }
                        statusEl.className = 'response-box error';
                        statusEl.style.display = 'block';
                    }
                }
            } catch (_) {
                if (statusEl) {
                    statusEl.textContent = 'خطای ارتباط با سرور.';
                    statusEl.className = 'response-box error';
                    statusEl.style.display = 'block';
                }
            }
        });
    }

    if (btnStartDisable2FA) {
        btnStartDisable2FA.addEventListener('click', () => {
            init2FAModalOtpControllers();
            if (disable2faOtpController) disable2faOtpController.clear();
            if (disableSection2FA) disableSection2FA.style.display = 'block';
            if (initialActions2FA) initialActions2FA.style.display = 'none';
            setTimeout(() => {
                const pwdInput = document.getElementById('disable_2fa_pwd');
                if (pwdInput && !pwdInput.value) {
                    pwdInput.focus();
                } else if (disable2faOtpController) {
                    disable2faOtpController.focus();
                }
            }, 120);
        });
    }

    if (formDisable2FA) {
        formDisable2FA.addEventListener('submit', async (e) => {
            e.preventDefault();
            const pwdInput = document.getElementById('disable_2fa_pwd');
            const pwd = pwdInput ? pwdInput.value : '';
            const code = document.getElementById('disable_2fa_code').value.trim();
            const statusEl = document.getElementById('disable-2fa-status');

            if (!pwd || !pwd.trim()) {
                triggerPasswordError(pwdInput);
                if (statusEl) {
                    statusEl.textContent = 'لطفاً کلمه عبور جاری را وارد نمایید.';
                    statusEl.className = 'response-box error';
                    statusEl.style.display = 'block';
                }
                return;
            }

            if (!code || code.length < 6) {
                if (disable2faOtpController) disable2faOtpController.showError();
                if (statusEl) {
                    statusEl.textContent = 'لطفاً کد ۶ رقمی نرم‌افزار را به طور کامل وارد نمایید.';
                    statusEl.className = 'response-box error';
                    statusEl.style.display = 'block';
                }
                return;
            }

            try {
                const res = await adminFetch('/api/admin/2fa/disable', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ password: pwd, totp_code: code })
                });
                if (res.ok) {
                    if (disable2faOtpController) disable2faOtpController.showSuccess();
                    if (currentAdmin) currentAdmin.totp_enabled = false;
                    showToast('احراز هویت دومرحله‌ای غیرفعال شد', 4000, 'info');
                    reset2FAModalViews();
                    await refresh2FAStatus();
                } else {
                    const err = (await res.text()).trim();
                    const isPwdErr = err.includes('کلمه عبور') || err.includes('رمز عبور') || (res.status === 401 && !err.includes('منقضی') && !err.includes('Unauthorized'));
                    if (isPwdErr) {
                        triggerPasswordError(pwdInput);
                    } else {
                        if (disable2faOtpController) disable2faOtpController.showError();
                    }
                    if (statusEl) {
                        if (err === 'Unauthorized' || (res.status === 401 && !isPwdErr)) {
                            statusEl.textContent = 'نشست کاری شما منقضی شده است. در حال انتقال به صفحه ورود...';
                            setTimeout(() => location.reload(), 1500);
                        } else {
                            statusEl.textContent = err || (isPwdErr ? 'کلمه عبور جاری اشتباه است.' : 'اطلاعات وارد شده صحیح نمی‌باشد.');
                        }
                        statusEl.className = 'response-box error';
                        statusEl.style.display = 'block';
                    }
                }
            } catch (_) {
                if (statusEl) {
                    statusEl.textContent = 'خطای ارتباط با سرور.';
                    statusEl.className = 'response-box error';
                    statusEl.style.display = 'block';
                }
            }
        });
    }

    let allReports = [];
    let filteredReports = [];
    let currentPage = 1;
    const itemsPerPage = 25;
    const selectedReportIds = new Set();
    let currentViewingIndex = -1;
    let previousUnreadCount = -1;

    const searchTracking = document.getElementById('search-tracking');
    const filterType = document.getElementById('filter-type');
    const filterPriority = document.getElementById('filter-priority');
    const filterReportType = document.getElementById('filter-report-type');
    const filterMonth = document.getElementById('filter-month');
    const filterReadStatus = document.getElementById('filter-read-status');
    const clearFiltersBtn = document.getElementById('clearFiltersBtn');
    const btnExportCSV = document.getElementById('btnExportCSV');

    let currentSortDateDirection = 'desc';

    const CATEGORICAL_CONFIG = {
        priority: {
            selectId: 'filter-priority',
            defaultTitle: 'اولویت',
            prefix: 'اولویت: ',
            cycle: ['all', 'normal', 'urgent', 'critical'],
            labels: {
                normal: 'عادی',
                urgent: 'فوری',
                critical: 'آنی'
            }
        },
        category: {
            selectId: 'filter-type',
            defaultTitle: 'موضوع',
            prefix: 'موضوع: ',
            cycle: ['all', 'تخلفات مالی', 'مسائل امنیتی', 'تخلفات اداری', 'تخلفات آموزشی', 'پرسنلی', 'دانش آموزی', 'مدارس غیردولتی', 'آموزش مجازی', 'فضای مجازی', 'سایر'],
            labels: {
                'تخلفات مالی': 'مالی',
                'مسائل امنیتی': 'امنیتی',
                'تخلفات اداری': 'اداری',
                'تخلفات آموزشی': 'آموزشی',
                'پرسنلی': 'پرسنلی',
                'دانش آموزی': 'دانش‌آموزی',
                'مدارس غیردولتی': 'غیردولتی',
                'آموزش مجازی': 'مجازی',
                'فضای مجازی': 'فضای مجازی',
                'سایر': 'سایر',
                'مالی': 'مالی',
                'امنیتی': 'امنیتی'
            }
        },
        type: {
            selectId: 'filter-report-type',
            defaultTitle: 'نوع گزارش',
            prefix: 'نوع: ',
            cycle: ['all', 'anonymous', 'identified'],
            labels: {
                anonymous: 'ناشناس',
                identified: 'با هویت'
            }
        },
        status: {
            selectId: 'filter-read-status',
            defaultTitle: 'وضعیت بررسی',
            prefix: 'وضعیت: ',
            cycle: ['all', 'unread', 'read', 'responded', 'new_message', 'regional_response'],
            labels: {
                unread: 'در نوبت بررسی',
                read: 'بررسی شده',
                responded: 'دارای پاسخ',
                new_message: 'پیام جدید',
                regional_response: 'پاسخ حراست'
            }
        }
    };

    const ADMIN_FILTER_DESCRIPTIONS = {
        'filter-type': {
            'all': 'نمایش کلیه موضوعات و دسته‌بندی‌های اداری',
            'تخلفات مالی': 'رشوه، اختلاس، فاکتورسازی و سوءاستفاده مالی',
            'مسائل امنیتی': 'حفاظت اسناد، سرقت، تهدید و مسائل امنیتی',
            'تخلفات اداری': 'تخلفات آیین‌نامه‌ای، عدم پاسخگویی و اداری',
            'تخلفات آموزشی': 'بی‌عدالتی نمرات، تخلف آزمون و تقلب',
            'پرسنلی': 'مسائل پرسنلی، غیبت غیرموجه و رفتار سازمانی',
            'دانش آموزی': 'سوءرفتار، تنبیه بدنی و شکایات دانش‌آموزی',
            'مدارس غیردولتی': 'شهریه غیرمجاز و تخلفات مراکز غیردولتی',
            'آموزش مجازی': 'سامانه‌های آموزشی شاد و بسترهای آنلاین',
            'فضای مجازی': 'کانال‌ها، گروه‌ها و تخلفات در فضای مجازی',
            'سایر': 'سایر موضوعات و گزارش‌های متفرقه'
        },
        'filter-read-status': {
            'all': 'نمایش همه وضعیت‌های بررسی کارتابل',
            'new_message': 'گزارش‌های دارای پیام جدید از سمت کاربر',
            'regional_response': 'پاسخ حراست ارسالی توسط ناحیه یا منطقه',
            'unread': 'در نوبت بررسی اولیه (گزارش‌های ثبت‌شده جدید)',
            'read': 'بررسی‌شده و پرونده فعال در جریان اقدام',
            'responded': 'پاسخ نهایی ثبت و به کاربر ابلاغ شده'
        },
        'filter-priority': {
            'all': 'نمایش همه اولویت‌های زمانی رسیدگی',
            'normal': 'فرآیند استاندارد و نوبت عادی بررسی',
            'urgent': 'رسیدگی سریع و دارای اولویت زمانی',
            'critical': 'اقدام فوری و بسیار حیاتی حراستی'
        },
        'filter-report-type': {
            'all': 'نمایش گزارش‌های ناشناس و هویت‌دار',
            'anonymous': 'ثبت‌شده به‌صورت ناشناس و بدون مشخصات',
            'identified': 'دارای نام و مشخصات فردی گزارش‌دهنده'
        },
        'detail-priority-select': {
            'normal': 'فرآیند استاندارد و نوبت عادی رسیدگی',
            'urgent': 'رسیدگی سریع با اولویت زمانی بالا',
            'critical': 'اقدام فوری، اضطراری و بسیار ویژه حراست'
        },
        'audit-filter-action': {
            'all': 'نمایش کلیه اقدامات ثبت‌شده در سامانه',
            'LOGIN': 'ورود کاربر اداری به پنل کاربری',
            'LOGOUT': 'خروج رسمی کاربر اداری از سیستم',
            'VIEW_REPORT_DETAILS': 'مشاهده پرونده و جزییات گزارش',
            'RESPOND_REPORT': 'ثبت یا ویرایش پاسخ گزارش',
            'TOGGLE_READ_STATUS': 'تغییر وضعیت خوانده‌شده پرونده',
            'CREATE_ADMIN_USER': 'تعریف و ثبت کاربر جدید اداری',
            'TOGGLE_USER_ACTIVE': 'فعال یا مسدودسازی حساب کاربری',
            'RESET_PASSWORD_USER': 'تغییر رمز عبور توسط مدیر ارشد',
            'DELETE_ADMIN_USER': 'حذف کامل حساب کاربر اداری',
            'CHANGE_OWN_PASSWORD': 'تغییر رمز عبور شخصی کاربر',
            'REVOKE_SESSION': 'ابطال نشست امنیتی دستگاه',
            'REVOKE_OTHER_SESSIONS': 'ابطال کلیه نشست‌های فعال دیگر'
        },
        'audit-filter-user': {
            'all': 'همه کاربران ثبت‌کننده لاگ در سامانه'
        },
        'new_role': {
            'edare_kol': 'دسترسی کامل به پرونده‌های حراست کل استان',
            'navahi_manategh': 'دسترسی استعلام و ارسال پاسخ حراست منطقه',
            'superuser': 'مدیر ارشد با دسترسی کامل به کلیه بخش‌ها و کاربران'
        },
        'edit_role': {
            'edare_kol': 'دسترسی کامل به پرونده‌های حراست کل استان',
            'navahi_manategh': 'دسترسی استعلام و ارسال پاسخ حراست منطقه',
            'superuser': 'مدیر ارشد با دسترسی کامل به کلیه بخش‌ها و کاربران'
        },
        'edit_is_active': {
            'true': 'حساب کاربری فعال و دارای مجوز ورود به سیستم',
            'false': 'حساب غیرفعال و مسدود، قطع کامل دسترسی به پنل'
        }
    };

    const adminCustomSelectControllers = new Map();

    function syncAllAdminCustomSelects() {
        adminCustomSelectControllers.forEach(ctrl => {
            if (ctrl && typeof ctrl.sync === 'function') {
                ctrl.sync();
            }
        });
    }

    function initAdminCustomSelect(selectId, descriptions = {}) {
        const nativeSelect = document.getElementById(selectId);
        if (!nativeSelect) return null;

        const parentGroup = nativeSelect.closest('.filter-group') || nativeSelect.closest('.form-group') || nativeSelect.closest('.detail-priority-wrapper') || nativeSelect.parentElement;
        if (!parentGroup) return null;

        nativeSelect.classList.add('native-select-hidden');
        nativeSelect.setAttribute('tabindex', '-1');
        nativeSelect.setAttribute('aria-hidden', 'true');

        const existingCustom = parentGroup.querySelector(`.admin-custom-dropdown[data-target="${selectId}"]`);
        if (existingCustom) existingCustom.remove();

        const dropdown = document.createElement('div');
        dropdown.className = 'admin-custom-dropdown';
        dropdown.setAttribute('data-target', selectId);
        if (selectId === 'detail-priority-select') {
            dropdown.classList.add('dropdown-detail-priority');
        } else if (selectId.startsWith('audit-filter-')) {
            dropdown.classList.add('dropdown-audit-control');
        } else if (selectId === 'new_role' || selectId === 'edit_role' || selectId === 'edit_is_active') {
            dropdown.classList.add('dropdown-modal-control');
        }

        const trigger = document.createElement('button');
        trigger.type = 'button';
        trigger.className = 'custom-dropdown-trigger';
        trigger.setAttribute('aria-haspopup', 'listbox');
        trigger.setAttribute('aria-expanded', 'false');

        const triggerContent = document.createElement('div');
        triggerContent.className = 'trigger-content';

        const triggerLabel = document.createElement('span');
        triggerLabel.className = 'trigger-label';

        triggerContent.appendChild(triggerLabel);

        const triggerChevron = document.createElement('span');
        triggerChevron.className = 'trigger-chevron';
        triggerChevron.innerHTML = '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"></path></svg>';

        trigger.appendChild(triggerContent);
        trigger.appendChild(triggerChevron);

        const menu = document.createElement('div');
        menu.className = 'custom-dropdown-menu';
        menu.setAttribute('role', 'listbox');

        const options = Array.from(nativeSelect.options);
        options.forEach(opt => {
            const item = document.createElement('div');
            item.className = 'dropdown-item';
            item.setAttribute('role', 'option');
            item.setAttribute('tabindex', '-1');
            item.setAttribute('data-value', opt.value);

            const contentSide = document.createElement('div');
            contentSide.className = 'item-content-side';

            const itemLabel = document.createElement('span');
            itemLabel.className = 'item-label';
            itemLabel.textContent = opt.textContent;

            const descText = descriptions[opt.value] || '';
            const itemDesc = document.createElement('span');
            itemDesc.className = 'item-description';
            itemDesc.textContent = descText;

            contentSide.appendChild(itemLabel);
            if (descText) {
                contentSide.appendChild(itemDesc);
            }

            const checkIndicator = document.createElement('span');
            checkIndicator.className = 'item-check-indicator';
            checkIndicator.setAttribute('aria-hidden', 'true');
            checkIndicator.innerHTML = '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>';

            item.appendChild(contentSide);
            item.appendChild(checkIndicator);

            item.addEventListener('click', (e) => {
                e.stopPropagation();
                if (nativeSelect.value !== opt.value) {
                    nativeSelect.value = opt.value;
                    nativeSelect.dispatchEvent(new Event('change', { bubbles: true }));
                    if (typeof nativeSelect.onchange === 'function') {
                        nativeSelect.onchange();
                    }
                }
                syncState();
                closeMenu();
                trigger.focus();
            });

            item.addEventListener('keydown', (e) => {
                if (e.key === 'ArrowDown') {
                    e.preventDefault();
                    const next = item.nextElementSibling;
                    if (next && next.classList.contains('dropdown-item')) next.focus();
                } else if (e.key === 'ArrowUp') {
                    e.preventDefault();
                    const prev = item.previousElementSibling;
                    if (prev && prev.classList.contains('dropdown-item')) {
                        prev.focus();
                    } else {
                        trigger.focus();
                    }
                } else if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    item.click();
                }
            });

            menu.appendChild(item);
        });

        dropdown.appendChild(trigger);
        dropdown.appendChild(menu);
        parentGroup.appendChild(dropdown);

        function openMenu() {
            document.querySelectorAll('.admin-custom-dropdown.open').forEach(d => {
                if (d !== dropdown) {
                    d.classList.remove('open');
                    const t = d.querySelector('.custom-dropdown-trigger');
                    if (t) t.setAttribute('aria-expanded', 'false');
                    const pg = d.closest('.filter-group') || d.closest('.form-group') || d.closest('.detail-priority-wrapper') || d.parentElement;
                    if (pg) pg.classList.remove('has-open-dropdown');
                }
            });
            const fc = dropdown.closest('.filter-container') || dropdown.closest('.adv-audit-filters-grid') || dropdown.closest('.admin-modal-box');
            if (fc) fc.classList.add('has-open-dropdown');
            parentGroup.classList.add('has-open-dropdown');
            dropdown.classList.add('open');
            trigger.setAttribute('aria-expanded', 'true');
        }

        function closeMenu() {
            dropdown.classList.remove('open');
            trigger.setAttribute('aria-expanded', 'false');
            parentGroup.classList.remove('has-open-dropdown');
            const fc = dropdown.closest('.filter-container') || dropdown.closest('.adv-audit-filters-grid') || dropdown.closest('.admin-modal-box');
            if (fc && !fc.querySelector('.admin-custom-dropdown.open')) {
                fc.classList.remove('has-open-dropdown');
            }
        }

        function toggleMenu() {
            if (dropdown.classList.contains('open')) {
                closeMenu();
            } else {
                openMenu();
            }
        }

        trigger.addEventListener('click', (e) => {
            e.stopPropagation();
            toggleMenu();
        });

        trigger.addEventListener('keydown', (e) => {
            if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
                e.preventDefault();
                if (!dropdown.classList.contains('open')) {
                    openMenu();
                }
                const firstSelected = menu.querySelector('.dropdown-item.selected') || menu.querySelector('.dropdown-item');
                if (firstSelected) firstSelected.focus();
            }
        });

        function syncState() {
            const currentVal = nativeSelect.value;
            const currentOpt = nativeSelect.selectedOptions[0] || options.find(o => o.value === currentVal) || options[0];
            if (currentOpt) {
                triggerLabel.textContent = currentOpt.textContent;
                trigger.setAttribute('title', currentOpt.textContent);
            }

            if (currentVal && currentVal !== 'all' && currentVal !== 'normal' && currentVal !== 'edare_kol' && currentVal !== 'true') {
                dropdown.classList.add('has-filter');
            } else {
                dropdown.classList.remove('has-filter');
            }

            const items = menu.querySelectorAll('.dropdown-item');
            items.forEach(it => {
                if (it.getAttribute('data-value') === currentVal) {
                    it.classList.add('selected');
                    it.setAttribute('aria-selected', 'true');
                } else {
                    it.classList.remove('selected');
                    it.setAttribute('aria-selected', 'false');
                }
            });
        }

        syncState();

        nativeSelect.addEventListener('change', syncState);
        nativeSelect.addEventListener('sync-custom', syncState);

        const controller = {
            dropdown,
            trigger,
            menu,
            sync: syncState,
            close: closeMenu
        };

        adminCustomSelectControllers.set(selectId, controller);
        return controller;
    }

    function initAdminCustomFilters() {
        ['filter-type', 'filter-read-status', 'filter-priority', 'filter-report-type', 'detail-priority-select', 'audit-filter-action', 'audit-filter-user', 'new_role', 'edit_role', 'edit_is_active'].forEach(id => {
            if (document.getElementById(id)) {
                initAdminCustomSelect(id, ADMIN_FILTER_DESCRIPTIONS[id]);
            }
        });

        document.addEventListener('click', (e) => {
            if (!e.target.closest('.admin-custom-dropdown')) {
                document.querySelectorAll('.admin-custom-dropdown.open').forEach(d => {
                    d.classList.remove('open');
                    const t = d.querySelector('.custom-dropdown-trigger');
                    if (t) t.setAttribute('aria-expanded', 'false');
                    const pg = d.closest('.filter-group') || d.closest('.form-group') || d.closest('.detail-priority-wrapper') || d.parentElement;
                    if (pg) pg.classList.remove('has-open-dropdown');
                });
                document.querySelectorAll('.has-open-dropdown').forEach(el => {
                    if (!el.querySelector('.admin-custom-dropdown.open')) {
                        el.classList.remove('has-open-dropdown');
                    }
                });
            }
        });

        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                const openDropdown = document.querySelector('.admin-custom-dropdown.open');
                if (openDropdown) {
                    openDropdown.classList.remove('open');
                    const t = openDropdown.querySelector('.custom-dropdown-trigger');
                    if (t) {
                        t.setAttribute('aria-expanded', 'false');
                        t.focus();
                    }
                    document.querySelectorAll('.has-open-dropdown').forEach(el => {
                        el.classList.remove('has-open-dropdown');
                    });
                }
            }
        });
    }

    function updateTableHeadersFromFilters() {
        Object.entries(CATEGORICAL_CONFIG).forEach(([colKey, conf]) => {
            const th = document.getElementById(`th-${colKey}`);
            if (!th) return;
            const selectEl = document.getElementById(conf.selectId);
            const val = (selectEl && selectEl.value) ? selectEl.value : 'all';
            const titleEl = th.querySelector('.th-title');
            const iconEl = th.querySelector('.th-sort-icon');

            if (!val || val === 'all') {
                if (titleEl) titleEl.textContent = conf.defaultTitle;
                if (iconEl) iconEl.innerHTML = '';
                th.classList.remove('is-filtered');
            } else {
                const label = conf.labels[val] || val;
                if (titleEl) titleEl.textContent = `${conf.prefix}${label}`;
                if (iconEl) {
                    iconEl.innerHTML = '<svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" class="th-chevron-icon"><polyline points="6 9 12 15 18 9"/></svg>';
                }
                th.classList.add('is-filtered');
            }
        });

        const dateTh = document.getElementById('th-date');
        if (dateTh) {
            const dateIcon = dateTh.querySelector('.th-sort-icon');
            if (dateIcon) {
                if (currentSortDateDirection === 'asc') {
                    dateIcon.innerHTML = '<svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" class="th-chevron-icon"><polyline points="18 15 12 9 6 15"/></svg>';
                    dateTh.classList.add('is-sorted');
                    dateTh.setAttribute('data-tooltip', 'مرتب‌سازی: قدیمی‌ترین به جدیدترین (کلیک جهت تغییر به جدیدترین)');
                } else {
                    dateIcon.innerHTML = '<svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" class="th-chevron-icon"><polyline points="6 9 12 15 18 9"/></svg>';
                    dateTh.classList.add('is-sorted');
                    dateTh.setAttribute('data-tooltip', 'مرتب‌سازی: جدیدترین به قدیمی‌ترین (کلیک جهت تغییر به قدیمی‌ترین)');
                }
            }
        }
        syncAllAdminCustomSelects();
    }

    function initTableHeadersInteractions() {
        Object.entries(CATEGORICAL_CONFIG).forEach(([colKey, conf]) => {
            const th = document.getElementById(`th-${colKey}`);
            if (!th) return;

            const handleCycle = () => {
                const selectEl = document.getElementById(conf.selectId);
                if (!selectEl) return;
                const currentVal = selectEl.value || 'all';
                const cycle = conf.cycle;
                let idx = cycle.indexOf(currentVal);
                if (idx === -1) {
                    if (currentVal === 'مالی') idx = cycle.indexOf('تخلفات مالی');
                    else if (currentVal === 'امنیتی') idx = cycle.indexOf('مسائل امنیتی');
                    else idx = 0;
                }
                const nextVal = cycle[(idx + 1) % cycle.length];
                selectEl.value = nextVal;
                applyFilters(true);
                updateTableHeadersFromFilters();
            };

            th.addEventListener('click', handleCycle);
            th.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    handleCycle();
                }
            });
        });

        const dateTh = document.getElementById('th-date');
        if (dateTh) {
            const handleDateSort = () => {
                currentSortDateDirection = currentSortDateDirection === 'desc' ? 'asc' : 'desc';
                applyFilters(true);
                updateTableHeadersFromFilters();
            };
            dateTh.addEventListener('click', handleDateSort);
            dateTh.addEventListener('keydown', (e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault();
                    handleDateSort();
                }
            });
        }
    }

    function debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    function initActionSearchBar() {
        const wrapper = document.getElementById('actionSearchWrapper');
        const input = document.getElementById('search-tracking');
        const dropdown = document.getElementById('actionSearchDropdown');
        const list = document.getElementById('actionSearchList');
        const titleEl = document.getElementById('actionSearchDropdownTitle');
        const countEl = document.getElementById('actionSearchCount');
        const iconSearch = document.getElementById('actionIconSearch');
        const iconSend = document.getElementById('actionIconSend');
        const iconBox = document.getElementById('searchIconBox');
        const kbdBadge = document.getElementById('searchKbdBadge');
        const toggleBtn = document.getElementById('expandableSearchToggle');
        const clearBtn = document.getElementById('expandableSearchClear');
        const iconSearchToggle = toggleBtn ? toggleBtn.querySelector('.icon-search') : null;
        const iconCloseToggle = toggleBtn ? toggleBtn.querySelector('.icon-close') : null;

        if (!wrapper || !input || !dropdown || !list) return;

        let activeIndex = -1;
        let currentItems = [];

        function openSearch() {
            wrapper.classList.add('is-open');
            if (iconSearchToggle) iconSearchToggle.classList.add('d-none');
            if (iconCloseToggle) iconCloseToggle.classList.remove('d-none');
            if (toggleBtn) {
                toggleBtn.setAttribute('aria-label', 'بستن جستجو');
                toggleBtn.setAttribute('data-tooltip', 'بستن جستجو (Esc)');
            }
            if (input.value.trim().length > 0 && clearBtn) {
                clearBtn.classList.remove('d-none');
            }
            setTimeout(() => {
                input.focus();
            }, 60);
        }

        function closeSearch(clearValue = true) {
            if (clearValue && input.value) {
                input.value = '';
                if (typeof applyFilters === 'function') {
                    applyFilters(true);
                }
            }
            wrapper.classList.remove('is-open');
            if (iconSearchToggle) iconSearchToggle.classList.remove('d-none');
            if (iconCloseToggle) iconCloseToggle.classList.add('d-none');
            if (toggleBtn) {
                toggleBtn.setAttribute('aria-label', 'جستجو');
                toggleBtn.setAttribute('data-tooltip', 'جستجو در گزارش‌ها (Ctrl+K)');
            }
            if (clearBtn) clearBtn.classList.add('d-none');
            if (kbdBadge) kbdBadge.style.opacity = '1';
            closeDropdown();
            input.blur();
        }

        if (toggleBtn) {
            toggleBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                if (wrapper.classList.contains('is-open')) {
                    if (input.value.trim().length > 0) {
                        closeSearch(true);
                    } else {
                        closeSearch(false);
                    }
                } else {
                    openSearch();
                    openDropdown();
                }
            });
        }

        if (clearBtn) {
            clearBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                input.value = '';
                clearBtn.classList.add('d-none');
                if (kbdBadge) kbdBadge.style.opacity = '1';
                input.focus();
                computeAndRender();
                if (typeof applyFilters === 'function') {
                    applyFilters(true);
                }
            });
        }

        const defaultActions = [
            {
                id: 'act-unread',
                title: 'گزارش‌های بررسی‌نشده',
                desc: 'فیلتر موارد جدید و نیازمند بررسی در کارتابل',
                tag: 'کارتابل',
                short: 'Ctrl+1',
                iconBoxClass: 'action-icon-amber',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>',
                run: () => {
                    if (filterReadStatus) filterReadStatus.value = 'unread';
                    applyFilters(true);
                    updateTableHeadersFromFilters();
                }
            },
            {
                id: 'act-critical',
                title: 'اولویت آنی و فوری',
                desc: 'فیلتر موارد بسیار مهم با فوریت بالا',
                tag: 'اولویت',
                short: 'Ctrl+2',
                iconBoxClass: 'action-icon-rose',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>',
                run: () => {
                    if (filterPriority) filterPriority.value = 'critical';
                    applyFilters(true);
                    updateTableHeadersFromFilters();
                }
            },
            {
                id: 'act-export',
                title: 'دریافت خروجی اکسل و CSV',
                desc: 'دانلود پرونده گزارش‌های کارتابل جاری',
                tag: 'عملیات',
                short: 'Ctrl+E',
                iconBoxClass: 'action-icon-emerald',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>',
                run: () => {
                    if (btnExportCSV) btnExportCSV.click();
                }
            },
            {
                id: 'act-analytics',
                title: 'آمار و تحلیل داده‌ها',
                desc: 'مشاهده شاخص‌ها و تحلیل فراوانی گزارش‌ها',
                tag: 'ناوبری',
                short: 'Ctrl+A',
                iconBoxClass: 'action-icon-blue',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="20" x2="18" y2="10"></line><line x1="12" y1="20" x2="12" y2="4"></line><line x1="6" y1="20" x2="6" y2="14"></line></svg>',
                run: () => {
                    document.getElementById('tab-btn-analytics')?.click();
                }
            },
            {
                id: 'act-2fa',
                title: 'احراز هویت دومرحله‌ای (۲FA)',
                desc: 'مدیریت و پیکربندی امنیت ورود دو عاملی',
                tag: 'امنیت',
                short: 'Ctrl+S',
                iconBoxClass: 'action-icon-cyan',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path><path d="m9 12 2 2 4-4"></path></svg>',
                run: () => {
                    document.getElementById('btnOpen2FAModal')?.click();
                }
            },
            {
                id: 'act-shortcuts',
                title: 'راهنمای کلیدهای میانبر',
                desc: 'مشاهده کلیدهای دسترسی سریع کیبورد',
                tag: 'راهنما',
                short: '?',
                iconBoxClass: 'action-icon-purple',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>',
                run: () => {
                    openShortcutsModal();
                }
            },
            {
                id: 'act-clear',
                title: 'بازنشانی فیلترها',
                desc: 'حذف تمامی فیلترها و نمایش کلیه گزارش‌ها',
                tag: 'بازنشانی',
                short: 'ESC',
                iconBoxClass: 'action-icon-blue',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"></polyline><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"></path></svg>',
                run: () => {
                    if (clearFiltersBtn) clearFiltersBtn.click();
                }
            }
        ];

        function normalizeQuery(str) {
            if (!str) return '';
            const fa2en = { '۰': '0', '۱': '1', '۲': '2', '۳': '3', '۴': '4', '۵': '5', '۶': '6', '۷': '7', '۸': '8', '۹': '9' };
            const ar2en = { '٠': '0', '١': '1', '٢': '2', '٣': '3', '٤': '4', '٥': '5', '٦': '6', '٧': '7', '٨': '8', '٩': '9' };
            return String(str).replace(/[۰-۹]/g, c => fa2en[c]).replace(/[٠-٩]/g, c => ar2en[c]).toLowerCase().trim();
        }

        function renderItems(items, title) {
            currentItems = items;
            activeIndex = -1;
            list.innerHTML = '';
            if (titleEl) titleEl.textContent = title;
            if (countEl) countEl.textContent = toPersianDigits(items.length);

            if (items.length === 0) {
                const emptyDiv = document.createElement('div');
                emptyDiv.className = 'action-search-empty';
                emptyDiv.innerHTML = '<svg class="action-empty-icon" viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line><line x1="8" y1="11" x2="14" y2="11"></line></svg><span>موردی برای نمایش یافت نشد</span>';
                list.appendChild(emptyDiv);
                return;
            }

            items.forEach((item, idx) => {
                const row = document.createElement('div');
                row.className = 'action-search-item';
                row.setAttribute('role', 'option');
                row.setAttribute('id', `action-item-${idx}`);
                row.dataset.index = String(idx);

                const mainCol = document.createElement('div');
                mainCol.className = 'action-search-item-main';

                const iconBox = document.createElement('span');
                iconBox.className = `action-item-icon-box ${item.iconBoxClass || 'action-icon-blue'}`;
                iconBox.innerHTML = item.iconSvg || '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle></svg>';

                const texts = document.createElement('div');
                texts.className = 'action-item-texts';

                const title = document.createElement('span');
                title.className = 'action-item-title';
                title.textContent = item.title;

                const desc = document.createElement('span');
                desc.className = 'action-item-desc';
                desc.textContent = item.desc;

                texts.appendChild(title);
                texts.appendChild(desc);
                mainCol.appendChild(iconBox);
                mainCol.appendChild(texts);

                const meta = document.createElement('div');
                meta.className = 'action-item-meta';

                if (item.short) {
                    const kbd = document.createElement('kbd');
                    kbd.className = 'action-kbd-shortcut';
                    kbd.textContent = item.short;
                    meta.appendChild(kbd);
                }

                if (item.tag) {
                    const tag = document.createElement('span');
                    tag.className = 'action-end-tag';
                    tag.textContent = item.tag;
                    meta.appendChild(tag);
                }

                row.appendChild(mainCol);
                row.appendChild(meta);

                row.addEventListener('click', (e) => {
                    e.stopPropagation();
                    executeItem(item);
                });

                list.appendChild(row);
            });
        }

        function updateSelection() {
            const rows = list.querySelectorAll('.action-search-item');
            rows.forEach((r, idx) => {
                if (idx === activeIndex) {
                    r.classList.add('is-selected');
                    r.scrollIntoView({ block: 'nearest' });
                    input.setAttribute('aria-activedescendant', r.id);
                } else {
                    r.classList.remove('is-selected');
                }
            });
            if (activeIndex === -1) {
                input.removeAttribute('aria-activedescendant');
            }
        }

        function executeItem(item) {
            closeDropdown();
            if (typeof item.run === 'function') {
                item.run();
            }
        }

        function openDropdown() {
            dropdown.style.display = 'block';
            input.setAttribute('aria-expanded', 'true');
            computeAndRender();
        }

        function closeDropdown() {
            dropdown.style.display = 'none';
            input.setAttribute('aria-expanded', 'false');
            activeIndex = -1;
            updateSelection();
        }

        function computeAndRender() {
            const rawVal = input.value || '';
            const q = normalizeQuery(rawVal);

            if (rawVal.trim().length > 0) {
                if (iconSearch) iconSearch.classList.add('d-none');
                if (iconSend) iconSend.classList.remove('d-none');
                if (iconBox) iconBox.classList.add('is-active');
                if (kbdBadge) kbdBadge.style.opacity = '0';
                if (clearBtn) clearBtn.classList.remove('d-none');
            } else {
                if (iconSearch) iconSearch.classList.remove('d-none');
                if (iconSend) iconSend.classList.add('d-none');
                if (iconBox) iconBox.classList.remove('is-active');
                if (kbdBadge) kbdBadge.style.opacity = '1';
                if (clearBtn) clearBtn.classList.add('d-none');
            }

            if (!q) {
                renderItems(defaultActions, 'دستورات و پیشنهادات سریع');
                return;
            }

            const matchedActions = defaultActions.filter(act => {
                const normTitle = normalizeQuery(act.title);
                const normDesc = normalizeQuery(act.desc);
                return normTitle.includes(q) || normDesc.includes(q);
            });

            const matchedReports = [];
            if (Array.isArray(allReports)) {
                for (let i = 0; i < allReports.length; i++) {
                    const r = allReports[i];
                    if (!r) continue;
                    const rTracking = normalizeQuery(r.tracking_code || r.id || '');
                    const rSubject = normalizeQuery(r.subject || '');
                    const rType = normalizeQuery(r.type || r.category || '');
                    const rMsg = normalizeQuery(r.message_text || '');

                    if (rTracking.includes(q) || rSubject.includes(q) || rType.includes(q) || rMsg.includes(q)) {
                        const priClass = r.priority === 'critical' ? 'action-icon-rose' : (r.priority === 'urgent' ? 'action-icon-amber' : 'action-icon-blue');
                        const priTag = r.priority === 'critical' ? 'آنی' : (r.priority === 'urgent' ? 'فوری' : 'عادی');
                        matchedReports.push({
                            id: `rep-${r.id}`,
                            title: `کد پیگیری: ${toPersianDigits(r.tracking_code || r.id)}`,
                            desc: `${r.subject || 'بدون موضوع'} · ${r.type || 'گزارش'}`,
                            tag: priTag,
                            short: 'نمایش',
                            iconBoxClass: priClass,
                            iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline></svg>',
                            run: () => {
                                viewReport(r.id);
                            }
                        });
                        if (matchedReports.length >= 6) break;
                    }
                }
            }

            const combined = [...matchedActions, ...matchedReports];
            renderItems(combined, `یافته‌ها برای «${rawVal.trim()}»`);
        }

        input.addEventListener('focus', () => {
            if (!wrapper.classList.contains('is-open')) openSearch();
            openDropdown();
        });

        input.addEventListener('click', () => {
            if (!wrapper.classList.contains('is-open')) openSearch();
            if (dropdown.style.display === 'none') openDropdown();
        });

        input.addEventListener('input', () => {
            computeAndRender();
        });

        input.addEventListener('keydown', (e) => {
            if (dropdown.style.display !== 'none' && currentItems.length > 0) {
                if (e.key === 'ArrowDown') {
                    e.preventDefault();
                    activeIndex = (activeIndex + 1) % currentItems.length;
                    updateSelection();
                    return;
                }
                if (e.key === 'ArrowUp') {
                    e.preventDefault();
                    activeIndex = (activeIndex - 1 + currentItems.length) % currentItems.length;
                    updateSelection();
                    return;
                }
                if (e.key === 'Enter') {
                    if (activeIndex >= 0 && activeIndex < currentItems.length) {
                        e.preventDefault();
                        executeItem(currentItems[activeIndex]);
                        return;
                    }
                    closeDropdown();
                    return;
                }
            }
            if (e.key === 'Escape') {
                e.preventDefault();
                closeDropdown();
                if (input.value.trim() !== '') {
                    input.value = '';
                    if (typeof applyFilters === 'function') applyFilters(true);
                }
                closeSearch(false);
            }
        });

        document.addEventListener('click', (e) => {
            if (!wrapper.contains(e.target)) {
                closeDropdown();
                if (wrapper.classList.contains('is-open') && input.value.trim() === '') {
                    closeSearch(false);
                }
            }
        });

        if (input.value.trim().length > 0) {
            wrapper.classList.add('is-open');
            if (iconSearchToggle) iconSearchToggle.classList.add('d-none');
            if (iconCloseToggle) iconCloseToggle.classList.remove('d-none');
            if (clearBtn) clearBtn.classList.remove('d-none');
            if (kbdBadge) kbdBadge.style.opacity = '0';
        }
    }

    if (searchTracking) {
        searchTracking.addEventListener('input', debounce(() => {
            requestAnimationFrame(() => applyFilters(true));
        }, 300));
    }
    if (filterType) filterType.addEventListener('change', () => { applyFilters(true); updateTableHeadersFromFilters(); });
    if (filterPriority) filterPriority.addEventListener('change', () => { applyFilters(true); updateTableHeadersFromFilters(); });
    if (filterReportType) filterReportType.addEventListener('change', () => { applyFilters(true); updateTableHeadersFromFilters(); });
    if (filterMonth) filterMonth.addEventListener('change', () => { applyFilters(true); updateTableHeadersFromFilters(); });
    if (filterReadStatus) filterReadStatus.addEventListener('change', () => { applyFilters(true); updateTableHeadersFromFilters(); });

    initAdminCustomFilters();
    initTableHeadersInteractions();
    updateTableHeadersFromFilters();
    initActionSearchBar();

    const refreshReportsBtn = document.getElementById('refreshReportsBtn');
    if (refreshReportsBtn) {
        refreshReportsBtn.addEventListener('click', async () => {
            const icon = refreshReportsBtn.querySelector('.refresh-icon');
            refreshReportsBtn.disabled = true;
            if (icon) icon.classList.add('spinning');

            try {
                await fetchReports();
                showToast('لیست گزارش‌ها بروزرسانی شد', 2500, 'success');
            } catch (err) {
                showToast('خطا در دریافت اطلاعات. ارتباط با سرور برقرار نشد.', 3500, 'error');
            } finally {
                setTimeout(() => {
                    if (icon) icon.classList.remove('spinning');
                    refreshReportsBtn.disabled = false;
                }, 400);
            }
        });
    }

    if (clearFiltersBtn) {
        clearFiltersBtn.addEventListener('click', () => {
            if (searchTracking) {
                searchTracking.value = '';
                const clearBtn = document.getElementById('expandableSearchClear');
                if (clearBtn) clearBtn.classList.add('d-none');
                const actionWrap = document.getElementById('actionSearchWrapper');
                if (actionWrap) {
                    actionWrap.classList.remove('is-open');
                    const toggleBtn = document.getElementById('expandableSearchToggle');
                    const iconS = toggleBtn ? toggleBtn.querySelector('.icon-search') : null;
                    const iconC = toggleBtn ? toggleBtn.querySelector('.icon-close') : null;
                    if (iconS) iconS.classList.remove('d-none');
                    if (iconC) iconC.classList.add('d-none');
                    if (toggleBtn) {
                        toggleBtn.setAttribute('aria-label', 'جستجو');
                        toggleBtn.setAttribute('data-tooltip', 'جستجو در گزارش‌ها (Ctrl+K)');
                    }
                }
            }
            if (filterType) filterType.value = 'all';
            if (filterPriority) filterPriority.value = 'all';
            if (filterReportType) filterReportType.value = 'all';
            if (filterMonth) filterMonth.value = 'all';
            if (filterReadStatus) filterReadStatus.value = 'all';
            currentSortDateDirection = 'desc';
            updateTableHeadersFromFilters();
            if (typeof window.resetShamsiDateRangePicker === 'function') {
                window.resetShamsiDateRangePicker();
            } else {
                applyFilters(true);
            }
        });
    }

    let activeShamsiDateRange = null;

    function toPersianDigits(n) {
        const f = ['۰', '۱', '۲', '۳', '۴', '۵', '۶', '۷', '۸', '۹'];
        return String(n).replace(/[0-9]/g, w => f[+w]);
    }

    function padZero(n) {
        return n < 10 ? '۰' + toPersianDigits(n) : toPersianDigits(n);
    }

    function jalaliToGregorian(jy, jm, jd) {
        jy += 1595;
        let days = -355668 + (365 * jy) + Math.floor(jy / 33) * 8 + Math.floor(((jy % 33) + 3) / 4) + jd + (jm < 7 ? (jm - 1) * 31 : ((jm - 7) * 30) + 186);
        let gy = 400 * Math.floor(days / 146097);
        days %= 146097;
        if (days > 36524) {
            gy += 100 * Math.floor(--days / 36524);
            days %= 36524;
            if (days >= 365) days++;
        }
        gy += 4 * Math.floor(days / 1461);
        days %= 1461;
        if (days > 365) {
            gy += Math.floor((days - 1) / 365);
            days = (days - 1) % 365;
        }
        const salA = [0, 31, ((gy % 4 === 0 && gy % 100 !== 0) || (gy % 400 === 0)) ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
        let gm = 0;
        while (gm < 13 && days >= salA[gm]) {
            days -= salA[gm];
            gm++;
        }
        return { gy, gm, gd: days + 1 };
    }

    function gregorianToJalali(gy, gm, gd) {
        const gD_m = [0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334];
        let gy2 = (gm > 2) ? (gy + 1) : gy;
        let days = 355666 + (365 * gy) + Math.floor((gy2 + 3) / 4) - Math.floor((gy2 + 99) / 100) + Math.floor((gy2 + 399) / 400) + gd + gD_m[gm - 1];
        let jy = -1595 + (33 * Math.floor(days / 12053));
        days %= 12053;
        jy += 4 * Math.floor(days / 1461);
        days %= 1461;
        if (days > 365) {
            jy += Math.floor((days - 1) / 365);
            days = (days - 1) % 365;
        }
        let jm = (days < 186) ? 1 + Math.floor(days / 31) : 7 + Math.floor((days - 186) / 30);
        let jd = 1 + ((days < 186) ? (days % 31) : ((days - 186) % 30));
        return { jy, jm, jd };
    }

    function getJalaliMonthDays(jy, jm) {
        if (jm <= 6) return 31;
        if (jm <= 11) return 30;
        const g = jalaliToGregorian(jy, 12, 30);
        const j = gregorianToJalali(g.gy, g.gm, g.gd);
        return (j.jm === 12 && j.jd === 30) ? 30 : 29;
    }

    function getJalaliDayStartMs(jy, jm, jd) {
        const g = jalaliToGregorian(jy, jm, jd);
        return new Date(g.gy, g.gm - 1, g.gd, 0, 0, 0, 0).getTime();
    }

    function getJalaliDayEndMs(jy, jm, jd) {
        const g = jalaliToGregorian(jy, jm, jd);
        return new Date(g.gy, g.gm - 1, g.gd, 23, 59, 59, 999).getTime();
    }

    function createShamsiDateRangePicker(options) {
        const trigger = document.getElementById(options.triggerId);
        const popover = document.getElementById(options.popoverId);
        const wrapper = document.getElementById(options.wrapperId);
        const filterContainer = options.filterContainerId ? document.getElementById(options.filterContainerId) : (trigger ? trigger.closest('.filter-container') : null);
        const textSpan = document.getElementById(options.textSpanId);
        const clearBtn = document.getElementById(options.clearBtnId);
        const prevBtn = document.getElementById(options.prevBtnId);
        const nextBtn = document.getElementById(options.nextBtnId);
        const titleEl = document.getElementById(options.titleElId);
        const gridEl = document.getElementById(options.gridElId);
        const resetBtn = document.getElementById(options.resetBtnId);
        const confirmBtn = document.getElementById(options.confirmBtnId);
        const presetChips = wrapper ? wrapper.querySelectorAll('.preset-chip') : [];

        if (!trigger || !popover || !gridEl) return null;

        const JALALI_MONTH_NAMES = [
            'فروردین', 'اردیبهشت', 'خرداد', 'تیر', 'مرداد', 'شهریور',
            'مهر', 'آبان', 'آذر', 'دی', 'بهمن', 'اسفند'
        ];

        const todayDate = getServerNow();
        const computedTodayKey = getJalaliDateKeyFromDate(todayDate);
        const fallbackTodayJ = gregorianToJalali(todayDate.getFullYear(), todayDate.getMonth() + 1, todayDate.getDate());
        const todayJy = computedTodayKey ? Math.floor(computedTodayKey / 10000) : fallbackTodayJ.jy;
        const todayJm = computedTodayKey ? Math.floor((computedTodayKey % 10000) / 100) : fallbackTodayJ.jm;
        const todayJd = computedTodayKey ? (computedTodayKey % 100) : fallbackTodayJ.jd;
        const todayKey = (todayJy * 10000) + (todayJm * 100) + todayJd;
        const todayJ = { jy: todayJy, jm: todayJm, jd: todayJd, key: todayKey };

        let currentViewJy = todayJ.jy;
        let currentViewJm = todayJ.jm;
        let pickingState = false;
        let tempStart = null;
        let tempEnd = null;
        let hoveredDay = null;

        function renderCalendar() {
            if (!titleEl || !gridEl) return;
            titleEl.textContent = JALALI_MONTH_NAMES[currentViewJm - 1] + ' ' + toPersianDigits(currentViewJy);
            gridEl.innerHTML = '';

            const daysInMonth = getJalaliMonthDays(currentViewJy, currentViewJm);
            const firstDayG = jalaliToGregorian(currentViewJy, currentViewJm, 1);
            const firstDayObj = new Date(firstDayG.gy, firstDayG.gm - 1, firstDayG.gd);
            const startWeekday = (firstDayObj.getDay() + 1) % 7;

            const prevMonth = currentViewJm === 1 ? 12 : currentViewJm - 1;
            const prevYear = currentViewJm === 1 ? currentViewJy - 1 : currentViewJy;
            const prevMonthDays = getJalaliMonthDays(prevYear, prevMonth);

            for (let i = startWeekday - 1; i >= 0; i--) {
                const dayNum = prevMonthDays - i;
                const cell = document.createElement('div');
                cell.className = 'cal-day-cell outside';
                cell.textContent = toPersianDigits(dayNum);
                cell.addEventListener('click', (e) => {
                    e.stopPropagation();
                    currentViewJm = prevMonth;
                    currentViewJy = prevYear;
                    renderCalendar();
                });
                gridEl.appendChild(cell);
            }

            const startKey = tempStart ? tempStart.key : null;
            let endKey = tempEnd ? tempEnd.key : null;
            if (pickingState && tempStart && hoveredDay) {
                endKey = hoveredDay.key >= tempStart.key ? hoveredDay.key : null;
            }

            let rangeMinKey = null;
            let rangeMaxKey = null;
            if (startKey && endKey) {
                rangeMinKey = Math.min(startKey, endKey);
                rangeMaxKey = Math.max(startKey, endKey);
            } else if (startKey) {
                rangeMinKey = startKey;
                rangeMaxKey = tempStart.key;
            }

            for (let d = 1; d <= daysInMonth; d++) {
                const cell = document.createElement('div');
                cell.className = 'cal-day-cell';
                cell.textContent = toPersianDigits(d);

                const cellKey = (currentViewJy * 10000) + (currentViewJm * 100) + d;

                if (cellKey === todayJ.key) {
                    cell.classList.add('today');
                }

                if (rangeMinKey !== null && rangeMaxKey !== null) {
                    if (cellKey === rangeMinKey && cellKey === rangeMaxKey) {
                        cell.classList.add('range-single');
                    } else if (cellKey === rangeMinKey) {
                        cell.classList.add('range-start');
                    } else if (cellKey === rangeMaxKey) {
                        cell.classList.add('range-end');
                    } else if (cellKey > rangeMinKey && cellKey < rangeMaxKey) {
                        cell.classList.add('range-middle');
                    }
                }

                cell.addEventListener('mouseenter', () => {
                    if (pickingState && tempStart) {
                        hoveredDay = {
                            jy: currentViewJy,
                            jm: currentViewJm,
                            jd: d,
                            key: cellKey
                        };
                        renderCalendarHighlights();
                    }
                });

                cell.addEventListener('click', (e) => {
                    e.stopPropagation();
                    if (!pickingState || !tempStart) {
                        tempStart = {
                            jy: currentViewJy,
                            jm: currentViewJm,
                            jd: d,
                            key: cellKey
                        };
                        tempEnd = null;
                        pickingState = true;
                        hoveredDay = null;
                        if (textSpan) {
                            textSpan.textContent = toPersianDigits(tempStart.jy) + '/' + padZero(tempStart.jm) + '/' + padZero(tempStart.jd) + ' تا ...';
                            textSpan.classList.add('has-range');
                        }
                        if (clearBtn) clearBtn.style.display = 'inline-flex';
                        renderCalendar();
                    } else {
                        const secondDay = {
                            jy: currentViewJy,
                            jm: currentViewJm,
                            jd: d,
                            key: cellKey
                        };
                        if (secondDay.key < tempStart.key) {
                            showToast('تاریخ پایان نمی‌تواند قبل از تاریخ شروع باشد', 3000, 'warning');
                            return;
                        }
                        tempEnd = secondDay;
                        pickingState = false;
                        hoveredDay = null;
                        commitDateRange();
                    }
                });

                cell.addEventListener('dblclick', (e) => {
                    e.stopPropagation();
                    const clickedDay = {
                        jy: currentViewJy,
                        jm: currentViewJm,
                        jd: d,
                        key: cellKey
                    };
                    if (pickingState && tempStart) {
                        if (clickedDay.key < tempStart.key) {
                            showToast('تاریخ پایان نمی‌تواند قبل از تاریخ شروع باشد', 3000, 'warning');
                            return;
                        }
                        tempEnd = clickedDay;
                    } else {
                        tempStart = clickedDay;
                        tempEnd = clickedDay;
                    }
                    pickingState = false;
                    hoveredDay = null;
                    commitDateRange();
                    closePopover();
                });

                gridEl.appendChild(cell);
            }

            const totalRendered = startWeekday + daysInMonth;
            const remaining = (7 - (totalRendered % 7)) % 7;
            const nextMonth = currentViewJm === 12 ? 1 : currentViewJm + 1;
            const nextYear = currentViewJm === 12 ? currentViewJy + 1 : currentViewJy;

            for (let i = 1; i <= remaining; i++) {
                const cell = document.createElement('div');
                cell.className = 'cal-day-cell outside';
                cell.textContent = toPersianDigits(i);
                cell.addEventListener('click', (e) => {
                    e.stopPropagation();
                    currentViewJm = nextMonth;
                    currentViewJy = nextYear;
                    renderCalendar();
                });
                gridEl.appendChild(cell);
            }
        }

        function renderCalendarHighlights() {
            if (!gridEl || !tempStart || !hoveredDay) return;
            const cells = gridEl.querySelectorAll('.cal-day-cell:not(.outside)');
            const isHoverEarlier = hoveredDay.key < tempStart.key;
            const minKey = tempStart.key;
            const maxKey = isHoverEarlier ? tempStart.key : hoveredDay.key;

            cells.forEach(cell => {
                const d = parseInt(cell.textContent.replace(/[۰-۹]/g, c => '۰۱۲۳۴۵۶۷۸۹'.indexOf(c)), 10);
                if (!d) return;
                const cellKey = (currentViewJy * 10000) + (currentViewJm * 100) + d;

                cell.classList.remove('range-start', 'range-end', 'range-middle', 'range-single', 'date-not-allowed');

                if (isHoverEarlier && cellKey === hoveredDay.key) {
                    cell.classList.add('date-not-allowed');
                }

                if (cellKey === minKey && cellKey === maxKey) {
                    cell.classList.add('range-single');
                } else if (cellKey === minKey) {
                    cell.classList.add('range-start');
                } else if (cellKey === maxKey) {
                    cell.classList.add('range-end');
                } else if (cellKey > minKey && cellKey < maxKey) {
                    cell.classList.add('range-middle');
                }
            });
        }

        function commitDateRange() {
            if (!tempStart) return;
            if (!tempEnd) {
                tempEnd = { ...tempStart };
            }
            const sKey = Math.min(tempStart.key, tempEnd.key);
            const eKey = Math.max(tempStart.key, tempEnd.key);
            const rangeObj = {
                startKey: sKey,
                endKey: eKey,
                startJ: tempStart,
                endJ: tempEnd
            };

            const sStr = toPersianDigits(tempStart.jy) + '/' + padZero(tempStart.jm) + '/' + padZero(tempStart.jd);
            const eStr = toPersianDigits(tempEnd.jy) + '/' + padZero(tempEnd.jm) + '/' + padZero(tempEnd.jd);
            if (textSpan) {
                textSpan.textContent = sStr + ' تا ' + eStr;
                textSpan.classList.add('has-range');
            }
            if (clearBtn) clearBtn.style.display = 'inline-flex';

            presetChips.forEach(c => c.classList.remove('active'));
            renderCalendar();
            if (typeof options.onCommit === 'function') {
                options.onCommit(rangeObj);
            }
        }

        function resetRange() {
            tempStart = null;
            tempEnd = null;
            pickingState = false;
            hoveredDay = null;
            if (textSpan) {
                textSpan.textContent = 'انتخاب بازه زمانی';
                textSpan.classList.remove('has-range');
            }
            if (clearBtn) clearBtn.style.display = 'none';

            presetChips.forEach(c => {
                c.classList.toggle('active', c.getAttribute('data-preset') === 'all');
            });

            currentViewJy = todayJ.jy;
            currentViewJm = todayJ.jm;
            renderCalendar();
            if (typeof options.onReset === 'function') {
                options.onReset();
            }
        }

        function closePopover() {
            popover.style.display = 'none';
            trigger.classList.remove('active');
            trigger.setAttribute('aria-expanded', 'false');
            if (wrapper) wrapper.classList.remove('popover-open');
            if (filterContainer) filterContainer.classList.remove('popover-open');
            const inboxSection = document.getElementById('reports-inbox-section');
            if (inboxSection) inboxSection.classList.remove('popover-active');
            const dashContainer = inboxSection ? inboxSection.closest('.dashboard-container') : null;
            if (dashContainer) dashContainer.classList.remove('popover-active');
        }

        function openPopover() {
            popover.style.display = 'flex';
            trigger.classList.add('active');
            trigger.setAttribute('aria-expanded', 'true');
            if (wrapper) wrapper.classList.add('popover-open');
            if (filterContainer) filterContainer.classList.add('popover-open');

            const inboxSection = document.getElementById('reports-inbox-section');
            if (inboxSection) inboxSection.classList.add('popover-active');
            const dashContainer = inboxSection ? inboxSection.closest('.dashboard-container') : null;
            if (dashContainer) dashContainer.classList.add('popover-active');

            popover.style.top = 'calc(100% + 8px)';
            popover.style.bottom = 'auto';

            const rect = popover.getBoundingClientRect();
            if (rect.left < 10) {
                popover.style.right = 'auto';
                popover.style.left = '0';
            } else {
                popover.style.right = '0';
                popover.style.left = 'auto';
            }
            renderCalendar();
        }

        popover.addEventListener('click', (e) => {
            e.stopPropagation();
        });

        trigger.addEventListener('click', (e) => {
            if (e.target === clearBtn) return;
            const isOpen = popover.style.display === 'flex';
            if (isOpen) {
                closePopover();
            } else {
                openPopover();
            }
        });

        if (clearBtn) {
            clearBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                resetRange();
            });
        }

        if (prevBtn) {
            prevBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                if (currentViewJm === 1) {
                    currentViewJm = 12;
                    currentViewJy--;
                } else {
                    currentViewJm--;
                }
                renderCalendar();
            });
        }

        if (nextBtn) {
            nextBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                if (currentViewJm === 12) {
                    currentViewJm = 1;
                    currentViewJy++;
                } else {
                    currentViewJm++;
                }
                renderCalendar();
            });
        }

        if (resetBtn) {
            resetBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                resetRange();
            });
        }

        if (confirmBtn) {
            confirmBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                closePopover();
            });
        }

        document.addEventListener('click', (e) => {
            if (popover.style.display !== 'flex') return;
            if (e.target && e.target.closest && (e.target.closest(`#${options.popoverId}`) || e.target.closest(`#${options.triggerId}`))) {
                return;
            }
            closePopover();
        });

        presetChips.forEach(chip => {
            chip.addEventListener('click', () => {
                const p = chip.getAttribute('data-preset');
                presetChips.forEach(c => c.classList.remove('active'));
                chip.classList.add('active');

                if (p === 'today') {
                    tempStart = {
                        jy: todayJ.jy,
                        jm: todayJ.jm,
                        jd: todayJ.jd,
                        key: todayJ.key
                    };
                    tempEnd = { ...tempStart };
                    currentViewJy = todayJ.jy;
                    currentViewJm = todayJ.jm;
                    commitDateRange();
                } else if (p === 'last7') {
                    const now = getServerNow();
                    const past = new Date(now.getTime() - 6 * 86400000);
                    const pastKey = getJalaliDateKeyFromDate(past) || todayJ.key;
                    const pastJy = Math.floor(pastKey / 10000);
                    const pastJm = Math.floor((pastKey % 10000) / 100);
                    const pastJd = pastKey % 100;
                    tempStart = {
                        jy: pastJy,
                        jm: pastJm,
                        jd: pastJd,
                        key: pastKey
                    };
                    tempEnd = {
                        jy: todayJ.jy,
                        jm: todayJ.jm,
                        jd: todayJ.jd,
                        key: todayJ.key
                    };
                    currentViewJy = todayJ.jy;
                    currentViewJm = todayJ.jm;
                    commitDateRange();
                } else if (p === 'last30') {
                    const now = getServerNow();
                    const past = new Date(now.getTime() - 29 * 86400000);
                    const pastKey = getJalaliDateKeyFromDate(past) || todayJ.key;
                    const pastJy = Math.floor(pastKey / 10000);
                    const pastJm = Math.floor((pastKey % 10000) / 100);
                    const pastJd = pastKey % 100;
                    tempStart = {
                        jy: pastJy,
                        jm: pastJm,
                        jd: pastJd,
                        key: pastKey
                    };
                    tempEnd = {
                        jy: todayJ.jy,
                        jm: todayJ.jm,
                        jd: todayJ.jd,
                        key: todayJ.key
                    };
                    currentViewJy = todayJ.jy;
                    currentViewJm = todayJ.jm;
                    commitDateRange();
                } else if (p === 'thisMonth') {
                    const daysInM = getJalaliMonthDays(todayJ.jy, todayJ.jm);
                    const startKey = (todayJ.jy * 10000) + (todayJ.jm * 100) + 1;
                    const endKey = (todayJ.jy * 10000) + (todayJ.jm * 100) + daysInM;
                    tempStart = {
                        jy: todayJ.jy,
                        jm: todayJ.jm,
                        jd: 1,
                        key: startKey
                    };
                    tempEnd = {
                        jy: todayJ.jy,
                        jm: todayJ.jm,
                        jd: daysInM,
                        key: endKey
                    };
                    currentViewJy = todayJ.jy;
                    currentViewJm = todayJ.jm;
                    commitDateRange();
                } else if (p === 'all') {
                    resetRange();
                }
            });
        });

        renderCalendar();

        return {
            resetRange,
            closePopover,
            openPopover
        };
    }

    const reportsDatePicker = createShamsiDateRangePicker({
        triggerId: 'dateRangeTrigger',
        popoverId: 'dateRangePopover',
        wrapperId: 'dateRangePickerWrapper',
        filterContainerId: 'reportsFilterContainer',
        textSpanId: 'dateRangeText',
        clearBtnId: 'dateRangeClear',
        prevBtnId: 'calPrevMonth',
        nextBtnId: 'calNextMonth',
        titleElId: 'calMonthTitle',
        gridElId: 'calDaysGrid',
        resetBtnId: 'calResetBtn',
        confirmBtnId: 'calConfirmBtn',
        onCommit: (range) => {
            activeShamsiDateRange = range;
            applyFilters(true);
        },
        onReset: () => {
            activeShamsiDateRange = null;
            applyFilters(true);
        }
    });

    if (reportsDatePicker) {
        window.resetShamsiDateRangePicker = reportsDatePicker.resetRange;
    }

    let auditActiveShamsiDateRange = null;

    const auditDatePicker = createShamsiDateRangePicker({
        triggerId: 'auditDateRangeTrigger',
        popoverId: 'auditDateRangePopover',
        wrapperId: 'auditDateRangePickerWrapper',
        filterContainerId: 'auditFilterContainer',
        textSpanId: 'auditDateRangeText',
        clearBtnId: 'auditDateRangeClear',
        prevBtnId: 'auditCalPrevMonth',
        nextBtnId: 'auditCalNextMonth',
        titleElId: 'auditCalMonthTitle',
        gridElId: 'auditCalDaysGrid',
        resetBtnId: 'auditCalResetBtn',
        confirmBtnId: 'auditCalConfirmBtn',
        onCommit: (range) => {
            auditActiveShamsiDateRange = range;
            applyAuditFiltersAndPagination(true);
        },
        onReset: () => {
            auditActiveShamsiDateRange = null;
            applyAuditFiltersAndPagination(true);
        }
    });

    if (auditDatePicker) {
        window.resetAuditShamsiDateRangePicker = auditDatePicker.resetRange;
    }

    function exportReportsToCSV() {
        if (!filteredReports || filteredReports.length === 0) {
            showToast('گزارشی برای خروجی اکسل وجود ندارد');
            return;
        }

        const headers = ['کد پیگیری', 'اولویت', 'موضوع', 'نوع گزارش', 'تاریخ ثبت', 'وضعیت بررسی', 'دارای پاسخ'];
        const rows = filteredReports.map(r => {
            const prioFa = r.priority === 'critical' ? 'آنی' : (r.priority === 'urgent' ? 'فوری' : 'عادی');
            let statusFa = 'در نوبت بررسی';
            const hasResp = Boolean(r.has_response || (r.admin_response && r.admin_response.trim() !== ''));
            if (r.has_new_message) {
                statusFa = 'پیام جدید';
            } else if (r.has_regional_response) {
                statusFa = 'پاسخ حراست';
            } else if (hasResp) {
                statusFa = 'پاسخ داده شده';
            } else if (r.is_read) {
                statusFa = 'بررسی شده';
            }
            const hasRespFa = hasResp ? 'بله' : 'خیر';
            const dateFa = formatJalaliDateTime(r.created_at);

            return [
                `"${r.tracking_code}"`,
                `"${prioFa}"`,
                `"${r.category || ''}"`,
                `"${typeFa}"`,
                `"${dateFa}"`,
                `"${statusFa}"`,
                `"${hasRespFa}"`
            ].join(',');
        });

        const csvContent = '\uFEFF' + [headers.join(','), ...rows].join('\r\n');
        const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.setAttribute('href', url);
        link.setAttribute('download', `Herasat_Reports_Export_${Date.now()}.csv`);
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        showToast('فایل اکسل با موفقیت دانلود شد');
    }

    if (btnExportCSV) {
        btnExportCSV.addEventListener('click', exportReportsToCSV);
    }

    function handleKPICardFilter(statusFilterValue) {

        if (reportsInboxSection) reportsInboxSection.style.display = 'block';
        if (detailSection) detailSection.style.display = 'none';
        if (listSection) listSection.style.display = 'block';
        if (filterContainer) filterContainer.style.display = '';
        reportsInboxSection?.classList.remove('in-detail-view');
        if (advancedSection) advancedSection.style.display = 'none';
        if (analyticsSection) analyticsSection.style.display = 'none';

        if (tabBtnReports) tabBtnReports.classList.add('active');
        if (tabBtnAnalytics) tabBtnAnalytics.classList.remove('active');
        if (tabBtnAdvanced) tabBtnAdvanced.classList.remove('active');

        if (filterReadStatus) filterReadStatus.value = statusFilterValue;
        applyFilters(true);
        updateTableHeadersFromFilters();

        const targetEl = document.querySelector('.filter-container') || listSection;
        if (targetEl) {
            targetEl.scrollIntoView({ behavior: getScrollBehavior(), block: 'start' });
        }
    }

    document.getElementById('kpi-total-card')?.addEventListener('click', () => handleKPICardFilter('all'));
    document.getElementById('kpi-unread-card')?.addEventListener('click', () => handleKPICardFilter('unread'));
    document.getElementById('kpi-read-card')?.addEventListener('click', () => handleKPICardFilter('read'));
    document.getElementById('kpi-responded-card')?.addEventListener('click', () => handleKPICardFilter('responded'));

    const paginationControls = document.getElementById('pagination-controls');
    if (paginationControls && !paginationControls._delegated) {
        paginationControls._delegated = true;
        paginationControls.addEventListener('click', (e) => {
            const btn = e.target.closest('.pagination-btn');
            if (!btn || btn.disabled || btn.getAttribute('aria-disabled') === 'true') return;

            const action = btn.getAttribute('data-action');
            const pageStr = btn.getAttribute('data-page');

            const totalItems = filteredReports.length;
            const totalPages = Math.ceil(totalItems / itemsPerPage) || 1;

            if (action === 'prev') {
                if (currentPage > 1) {
                    currentPage--;
                    renderCurrentPage();
                }
            } else if (action === 'next') {
                if (currentPage < totalPages) {
                    currentPage++;
                    renderCurrentPage();
                }
            } else if (pageStr) {
                const targetPage = parseInt(pageStr, 10);
                if (!isNaN(targetPage) && targetPage >= 1 && targetPage <= totalPages && targetPage !== currentPage) {
                    currentPage = targetPage;
                    renderCurrentPage();
                }
            }
        });
    }

    window.addEventListener('resize', debounce(() => {
        if (filteredReports && filteredReports.length > 0) {
            const totalItems = filteredReports.length;
            const totalPages = Math.ceil(totalItems / itemsPerPage) || 1;
            renderPaginationControls(currentPage, totalPages);
        }
    }, 200));

    function parseTimestamp(isoStr) {
        if (!isoStr) return null;

        const hasTz = /Z$|[+-]\d{2}:?\d{2}$/.test(isoStr);
        const d = new Date(hasTz ? isoStr : isoStr + 'Z');
        return isNaN(d.getTime()) ? null : d;
    }

    const cachedJalaliFormatterFull = new Intl.DateTimeFormat('fa-IR', {
        timeZone: 'Asia/Tehran',
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
    });

    const cachedJalaliFormatterShort = new Intl.DateTimeFormat('fa-IR', {
        timeZone: 'Asia/Tehran',
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
    });

    const cachedJalaliMonthFormatter = new Intl.DateTimeFormat('fa-IR-u-nu-latn', {
        timeZone: 'Asia/Tehran',
        month: '2-digit'
    });

    const cachedJalaliPartsFormatter = new Intl.DateTimeFormat('fa-IR-u-nu-latn', {
        timeZone: 'Asia/Tehran',
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
    });

    function getJalaliDateKeyFromDate(dateObj) {
        if (!dateObj) return 0;
        try {
            const parts = cachedJalaliPartsFormatter.formatToParts(dateObj);
            let y = 0, m = 0, d = 0;
            for (let i = 0; i < parts.length; i++) {
                if (parts[i].type === 'year') y = parseInt(parts[i].value, 10);
                else if (parts[i].type === 'month') m = parseInt(parts[i].value, 10);
                else if (parts[i].type === 'day') d = parseInt(parts[i].value, 10);
            }
            return (y * 10000) + (m * 100) + d;
        } catch (_) {
            return 0;
        }
    }

    function formatJalaliDateTime(isoStr, includeSeconds = false) {
        if (!isoStr) return 'نامشخص';
        try {
            const dateObj = parseTimestamp(isoStr);
            if (!dateObj) return 'نامشخص';
            return (includeSeconds ? cachedJalaliFormatterFull : cachedJalaliFormatterShort).format(dateObj);
        } catch (e) {
            return 'نامشخص';
        }
    }

    function indexReportForSearch(report) {
        if (!report) return;
        report._normTracking = normalizePersianForSearch(report.tracking_code || '');
        report._normCategory = normalizePersianForSearch(report.category || '');
        report._normDesc = normalizePersianForSearch(report.description || '');
        report._lowerTracking = (report.tracking_code || '').toLowerCase();
        report._lowerCategory = (report.category || '').toLowerCase();
        report._lowerDesc = (report.description || '').toLowerCase();

        if (report.created_at) {
            try {
                const d = parseTimestamp(report.created_at);
                if (d) {
                    report._shamsiMonth = cachedJalaliMonthFormatter.format(d);
                    report._timestamp = d.getTime();
                    report._jalaliDateKey = getJalaliDateKeyFromDate(d);
                } else {
                    report._shamsiMonth = '';
                    report._timestamp = 0;
                    report._jalaliDateKey = 0;
                }
            } catch (e) {
                report._shamsiMonth = '';
                report._timestamp = 0;
                report._jalaliDateKey = 0;
            }
        } else {
            report._shamsiMonth = '';
            report._timestamp = 0;
            report._jalaliDateKey = 0;
        }

        if (report.updated_at) {
            try {
                const ud = parseTimestamp(report.updated_at);
                report._updateTimestamp = ud ? ud.getTime() : report._timestamp;
            } catch (e) {
                report._updateTimestamp = report._timestamp;
            }
        } else {
            report._updateTimestamp = report._timestamp;
        }
    }

    let _elKPITotal = null;
    let _elKPIUnread = null;
    let _elKPIRead = null;
    let _elKPIResponded = null;
    let _elKPIPulse = null;

    function updateKPICards() {
        const total = allReports.length;
        let unread = 0;
        let responded = 0;
        let read = 0;

        for (let i = 0; i < total; i++) {
            const r = allReports[i];
            const hasResp = Boolean(r.has_response || (r.admin_response && r.admin_response.trim() !== ''));
            if (!r.is_read) {
                unread++;
            }
            if (hasResp) {
                responded++;
            } else if (r.is_read && !r.has_new_message && !(r.has_response || (r.admin_response && r.admin_response.trim() !== ''))) {
                read++;
            }
        }

        if (!_elKPITotal) _elKPITotal = document.getElementById('kpi-total-count');
        if (!_elKPIUnread) _elKPIUnread = document.getElementById('kpi-unread-count');
        if (!_elKPIRead) _elKPIRead = document.getElementById('kpi-read-count');
        if (!_elKPIResponded) _elKPIResponded = document.getElementById('kpi-responded-count');
        if (!_elKPIPulse) _elKPIPulse = document.getElementById('kpi-unread-pulse');

        if (_elKPITotal) _elKPITotal.textContent = toPersianNum(total);
        if (_elKPIUnread) _elKPIUnread.textContent = toPersianNum(unread);
        if (_elKPIRead) _elKPIRead.textContent = toPersianNum(read);
        if (_elKPIResponded) _elKPIResponded.textContent = toPersianNum(responded);

        if (_elKPIPulse) _elKPIPulse.style.display = unread > 0 ? 'block' : 'none';
    }

    const ESCAPE_MAP = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' };
    function escapeHTML(str) {
        if (str == null) return '';
        return String(str).replace(/[&<>"']/g, c => ESCAPE_MAP[c]);
    }

    function normalizePersianForSearch(str) {
        if (!str) return '';
        return String(str)
            .replace(/[\u200c\u200b\u200e\u200f\uFEFF]/g, '')
            .replace(/[يئ]/g, 'ی')
            .replace(/[ك]/g, 'ک')
            .replace(/[ة]/g, 'ه')
            .replace(/[آأإ]/g, 'ا')
            .replace(/[ؤ]/g, 'و')
            .replace(/[\u064B-\u065F\u0670]/g, '')
            .toLowerCase()
            .trim();
    }

    let _persianRegexCache = { term: null, regex: null };

    function makePersianRegex(term) {
        if (!term || !term.trim()) return null;
        const trimmed = term.trim().replace(/[\u200c\u200b\u200e\u200f\uFEFF]/g, '');
        if (!trimmed) return null;
        if (_persianRegexCache.term === trimmed) return _persianRegexCache.regex;

        let pattern = '';
        for (let i = 0; i < trimmed.length; i++) {
            const char = trimmed[i];
            let charPattern = char.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
            if (char === 'ی' || char === 'ي' || char === 'ئ') {
                charPattern = '[یييئ]';
            } else if (char === 'ک' || char === 'ك') {
                charPattern = '[کك]';
            } else if (char === 'ا' || char === 'آ' || char === 'أ' || char === 'إ') {
                charPattern = '[اآأإ]';
            } else if (char === 'ه' || char === 'ة') {
                charPattern = '[هة]';
            } else if (char === 'و' || char === 'ؤ') {
                charPattern = '[وؤ]';
            }
            pattern += (i > 0 ? '[\\u200c\\s]?' : '') + charPattern;
        }
        _persianRegexCache = { term: trimmed, regex: new RegExp(`(${pattern})`, 'gi') };
        return _persianRegexCache.regex;
    }

    function highlightSearchTerm(text, term, precompiledRegex = null) {
        if (!text) return 'ندارد';
        const escaped = escapeHTML(text);
        if (!term || !term.trim()) return escaped;
        const regex = precompiledRegex || makePersianRegex(term);
        if (!regex) return escaped;
        regex.lastIndex = 0;
        return escaped.replace(regex, '<span class="search-highlight">$1</span>');
    }

    function getMatchSnippet(fullText, term, maxLength = 80, precompiledRegex = null) {
        if (!fullText || !term || !term.trim()) return '';
        const regex = precompiledRegex || makePersianRegex(term);
        if (!regex) return '';
        regex.lastIndex = 0;
        const match = regex.exec(fullText);
        if (!match) return '';
        const matchIdx = match.index;
        const matchLen = match[0].length;

        const beforeLen = Math.floor((maxLength - matchLen) / 2);
        let start = Math.max(0, matchIdx - beforeLen);
        let end = Math.min(fullText.length, start + maxLength);
        if (end === fullText.length) {
            start = Math.max(0, end - maxLength);
        }

        let snippet = fullText.slice(start, end);
        if (start > 0) snippet = '...' + snippet;
        if (end < fullText.length) snippet = snippet + '...';

        return highlightSearchTerm(snippet, term, regex);
    }

    let _elBulkToolbar = null;
    let _elBulkCount = null;
    let _elSelectAllReports = null;

    function updateBulkToolbar() {
        if (!_elBulkToolbar) _elBulkToolbar = document.getElementById('bulkActionsToolbar');
        if (!_elBulkCount) _elBulkCount = document.getElementById('bulkSelectedCount');
        if (!_elSelectAllReports) _elSelectAllReports = document.getElementById('selectAllReports');

        const count = selectedReportIds.size;
        if (_elBulkCount) _elBulkCount.textContent = toPersianNum(count);
        if (_elBulkToolbar) _elBulkToolbar.style.display = count > 0 ? 'flex' : 'none';

        const btnBulkDelete = document.getElementById('btnBulkDelete');
        if (btnBulkDelete) {
            const isSuper = Boolean(currentAdmin && (currentAdmin.is_superuser || currentAdmin.role === 'superuser'));
            btnBulkDelete.style.display = (count > 0 && isSuper) ? 'inline-flex' : 'none';
        }

        const startIndex = (currentPage - 1) * itemsPerPage;
        const endIndex = Math.min(startIndex + itemsPerPage, filteredReports.length);
        const currentPageItems = filteredReports.slice(startIndex, endIndex);

        if (_elSelectAllReports) {
            if (currentPageItems.length > 0 && currentPageItems.every(r => selectedReportIds.has(r.id))) {
                _elSelectAllReports.checked = true;
                _elSelectAllReports.indeterminate = false;
            } else if (currentPageItems.some(r => selectedReportIds.has(r.id))) {
                _elSelectAllReports.checked = false;
                _elSelectAllReports.indeterminate = true;
            } else {
                _elSelectAllReports.checked = false;
                _elSelectAllReports.indeterminate = false;
            }
        }
    }

    document.getElementById('selectAllReports')?.addEventListener('change', (e) => {
        const startIndex = (currentPage - 1) * itemsPerPage;
        const endIndex = Math.min(startIndex + itemsPerPage, filteredReports.length);
        const currentPageItems = filteredReports.slice(startIndex, endIndex);

        if (e.target.checked) {
            currentPageItems.forEach(r => selectedReportIds.add(r.id));
        } else {
            currentPageItems.forEach(r => selectedReportIds.delete(r.id));
        }
        renderCurrentPage();
        updateBulkToolbar();
    });

    document.getElementById('btnBulkCancel')?.addEventListener('click', () => {
        selectedReportIds.clear();
        renderCurrentPage();
        updateBulkToolbar();
    });

    document.getElementById('btnBulkDelete')?.addEventListener('click', async () => {
        if (!currentAdmin || (!currentAdmin.is_superuser && currentAdmin.role !== 'superuser')) {
            showToast('تنها پشتیبانی سامانه مجاز به حذف پرونده می‌باشد.', 3000, 'warning');
            return;
        }

        const ids = Array.from(selectedReportIds);
        if (ids.length === 0) return;

        const countFa = toPersianNum(ids.length);
        const confirmed = await showConfirmDialog({
            title: 'حذف گروهی پرونده‌ها',
            description: `اخطار امنیتی پشتیبانی سامانه: آیا از حذف دائمی ${countFa} پرونده انتخاب‌شده و کلیه فایل‌های ضمیمه آن‌ها از پایگاه داده و سرور اطمینان دارید؟ این عملیات غیرقابل بازگشت است و در لاگ‌های بازرسی ثبت می‌گردد.`,
            confirmText: `بله، حذف دائمی ${countFa} پرونده`,
            cancelText: 'انصراف',
            variant: 'danger',
            iconType: 'trash'
        });
        if (!confirmed) return;

        try {
            const response = await adminFetch('/api/admin/report/delete', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ report_ids: ids })
            });

            if (!response.ok) {
                const errText = await response.text();
                throw new Error(errText || 'خطا در حذف گروهی پرونده‌ها');
            }

            const data = await response.json();
            showToast(data.message || `${countFa} پرونده با موفقیت حذف گردید`, 3500, 'success');
            selectedReportIds.clear();
            updateBulkToolbar();
            await fetchReports().catch(() => { });
        } catch (err) {
            showToast('خطا در اجرای عملیات حذف گروهی: ' + err.message, 4000, 'error');
        }
    });

    document.getElementById('btnBulkMarkRead')?.addEventListener('click', async () => {
        if (selectedReportIds.size === 0) return;
        const unreadSelected = Array.from(selectedReportIds).filter(id => {
            const rep = allReports.find(r => r.id === id);
            return rep && !rep.is_read;
        });

        if (unreadSelected.length === 0) {
            showToast('تمامی موارد انتخابی قبلاً بررسی شده‌اند.');
            selectedReportIds.clear();
            updateBulkToolbar();
            return;
        }

        try {
            await Promise.all(unreadSelected.map(id =>
                adminFetch('/api/admin/toggle-read', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ report_id: id })
                })
            ));
            showToast(`${toPersianNum(unreadSelected.length)} پرونده به عنوان بررسی‌شده علامت‌گذاری شدند`, 2500, 'success');
            selectedReportIds.clear();
            updateBulkToolbar();
            await fetchReports().catch(() => { });
        } catch (err) {
            showToast('خطا در انجام عملیات گروهی', 3000, 'error');
        }
    });

    function renderReportsTableSkeleton(count = 8) {
        if (!reportsBody) return;
        let html = '';
        for (let i = 0; i < count; i++) {
            html += `
                <tr class="skeleton-tr" aria-hidden="true">
                    <td><div class="skeleton-shimmer skeleton-td-check"></div></td>
                    <td><div class="skeleton-shimmer skeleton-td-tracking"></div></td>
                    <td><div class="skeleton-shimmer skeleton-td-priority"></div></td>
                    <td><div class="skeleton-shimmer skeleton-td-date"></div></td>
                    <td><div class="skeleton-shimmer skeleton-td-category"></div></td>
                    <td><div class="skeleton-shimmer skeleton-td-type"></div></td>
                    <td><div class="skeleton-shimmer skeleton-td-status"></div></td>
                    <td><div class="skeleton-shimmer skeleton-td-actions"></div></td>
                </tr>
            `;
        }
        reportsBody.innerHTML = html;
    }

    let lastReportsSignature = '';

    async function fetchReports(forceRefresh = false, showSkeleton = false) {
        if (!dashboardSection || dashboardSection.style.display === 'none' || dashboardSection.classList.contains('d-none') || !currentAdmin) {
            return false;
        }
        if (showSkeleton || !allReports || allReports.length === 0) {
            renderReportsTableSkeleton(8);
        }
        try {
            const queryParams = new URLSearchParams();
            queryParams.set('all', 'true');
            const reportsUrl = `/api/admin/reports?${queryParams.toString()}`;
            const response = await adminFetch(reportsUrl);
            if (response.status === 401) {
                forceLogoutOnConnectionDrop();
                return false;
            }
            if (!response.ok) {
                throw new Error(`خطای دریافت اطلاعات: ${response.status}`);
            }
            const data = await response.json();

            let fetchedReports = [];
            if (Array.isArray(data)) {
                fetchedReports = data;
            } else if (data && Array.isArray(data.reports)) {
                fetchedReports = data.reports;
            } else {
                fetchedReports = [];
            }

            let currentUnread = 0;
            for (let i = 0; i < fetchedReports.length; i++) {
                if (!fetchedReports[i].is_read) currentUnread++;
            }

            if (previousUnreadCount !== -1 && currentUnread > previousUnreadCount) {
                const diff = currentUnread - previousUnreadCount;
                showToast(`${toPersianNum(diff)} گزارش جدید در کارتابل دریافت شد`, 4000, 'info');
            }
            previousUnreadCount = currentUnread;

            const firstReport = fetchedReports[0];
            const newSignature = `${fetchedReports.length}_${currentUnread}_${firstReport ? firstReport.id : 0}_${firstReport ? (firstReport.updated_at || firstReport.created_at) : ''}`;
            if (!forceRefresh && allReports.length > 0 && newSignature === lastReportsSignature) {
                return true;
            }
            lastReportsSignature = newSignature;
            allReports = fetchedReports;

            for (let i = 0; i < allReports.length; i++) {
                indexReportForSearch(allReports[i]);
            }

            updateKPICards();
            applyFilters();
            updateTableHeadersFromFilters();
            return true;
        } catch (error) {
            console.error("Failed to load reports:", error);
            if (reportsBody && (!allReports || allReports.length === 0)) {
                reportsBody.innerHTML = `<tr><td colspan="8" style="text-align:center; padding:30px; color:var(--color-danger, #ef4444);">خطا در دریافت اطلاعات گزارش‌ها: ${escapeHTML(error.message)}</td></tr>`;
            }
            throw error;
        }
    }

    let activeSearchContext = { rawTerm: '', normTerm: '', lowerTerm: '', regex: null };

    function applyFilters(forceResetPage) {
        if (forceResetPage) currentPage = 1;

        const rawTerm = searchTracking ? searchTracking.value.trim() : '';
        if (activeSearchContext.rawTerm !== rawTerm) {
            activeSearchContext = {
                rawTerm,
                normTerm: normalizePersianForSearch(rawTerm),
                lowerTerm: rawTerm.toLowerCase(),
                regex: rawTerm ? makePersianRegex(rawTerm) : null
            };
        }
        const { normTerm, lowerTerm } = activeSearchContext;
        const typeMode = (filterType && filterType.value) ? filterType.value : 'all';
        const priorityMode = filterPriority ? filterPriority.value : 'all';
        const reportTypeFilter = document.getElementById('filter-report-type');
        const reportTypeMode = reportTypeFilter ? reportTypeFilter.value : 'all';
        const targetMonth = filterMonth ? filterMonth.value : 'all';
        const readMode = filterReadStatus ? filterReadStatus.value : 'all';

        filteredReports = allReports.filter(report => {
            let matchSearch = true;
            if (normTerm) {
                matchSearch = (report._normTracking && report._normTracking.includes(normTerm)) ||
                    (report._normCategory && report._normCategory.includes(normTerm)) ||
                    (report._normDesc && report._normDesc.includes(normTerm)) ||
                    (report._lowerTracking && report._lowerTracking.includes(lowerTerm)) ||
                    (report._lowerCategory && report._lowerCategory.includes(lowerTerm)) ||
                    (report._lowerDesc && report._lowerDesc.includes(lowerTerm));
            }
            const matchType = typeMode === 'all' || report.category === typeMode || (report.category && (report.category.includes(typeMode) || typeMode.includes(report.category)));
            const matchPriority = priorityMode === 'all' || (report.priority || 'normal') === priorityMode;
            const matchReportType = reportTypeMode === 'all' || (report.report_type || 'identified') === reportTypeMode;
            const matchMonth = targetMonth === 'all' || report._shamsiMonth === targetMonth;

            let matchDate = true;
            if (activeShamsiDateRange && activeShamsiDateRange.startKey && activeShamsiDateRange.endKey) {
                const rKey = report._jalaliDateKey || (report.created_at ? getJalaliDateKeyFromDate(parseTimestamp(report.created_at)) : 0);
                matchDate = rKey >= activeShamsiDateRange.startKey && rKey <= activeShamsiDateRange.endKey;
            } else if (activeShamsiDateRange && activeShamsiDateRange.startKey) {
                const rKey = report._jalaliDateKey || (report.created_at ? getJalaliDateKeyFromDate(parseTimestamp(report.created_at)) : 0);
                matchDate = rKey >= activeShamsiDateRange.startKey;
            }

            let matchRead = true;
            const hasResp = Boolean(report.has_response === true || (report.admin_response && report.admin_response.trim() !== ''));
            if (readMode === 'new_message') matchRead = report.has_new_message === true;
            else if (readMode === 'regional_response') matchRead = Boolean(report.has_regional_response || report.has_any_regional_response);
            else if (readMode === 'read') matchRead = report.is_read === true && !hasResp && !report.has_new_message && !report.has_regional_response;
            else if (readMode === 'unread') matchRead = !report.is_read;
            else if (readMode === 'responded') matchRead = hasResp;

            return matchSearch && matchType && matchPriority && matchReportType && matchMonth && matchDate && matchRead;
        });

        filteredReports.sort((a, b) => {
            const timeA = a._timestamp || 0;
            const timeB = b._timestamp || 0;
            if (timeA !== timeB) {
                return currentSortDateDirection === 'asc' ? (timeA - timeB) : (timeB - timeA);
            }
            return currentSortDateDirection === 'asc' ? ((a.id || 0) - (b.id || 0)) : ((b.id || 0) - (a.id || 0));
        });

        renderCurrentPage();
    }

    function calculatePagination({ currentPage, totalPages, paginationItemsToDisplay = 7 }) {
        if (totalPages <= paginationItemsToDisplay) {
            return {
                pages: Array.from({ length: totalPages }, (_, i) => i + 1),
                showLeftEllipsis: false,
                showRightEllipsis: false
            };
        }

        const showLeftEllipsis = currentPage - 1 > paginationItemsToDisplay / 2;
        const showRightEllipsis = totalPages - currentPage + 1 > paginationItemsToDisplay / 2;

        const halfDisplay = Math.floor(paginationItemsToDisplay / 2);
        const initialRange = {
            start: currentPage - halfDisplay,
            end: currentPage + halfDisplay,
        };

        const adjustedRange = {
            start: Math.max(1, initialRange.start),
            end: Math.min(totalPages, initialRange.end),
        };

        if (adjustedRange.start === 1) {
            adjustedRange.end = paginationItemsToDisplay;
        }
        if (adjustedRange.end === totalPages) {
            adjustedRange.start = totalPages - paginationItemsToDisplay + 1;
        }

        if (showLeftEllipsis) adjustedRange.start++;
        if (showRightEllipsis) adjustedRange.end--;

        const pages = Array.from(
            { length: adjustedRange.end - adjustedRange.start + 1 },
            (_, i) => adjustedRange.start + i
        );

        return {
            pages,
            showLeftEllipsis,
            showRightEllipsis,
        };
    }

    function renderPaginationControls(currPage, totalPgs) {
        const controls = document.getElementById('pagination-controls');
        if (!controls) return;

        if (totalPgs <= 1) {
            controls.innerHTML = '';
            return;
        }

        const isMobile = window.innerWidth < 640;
        const paginationItemsToDisplay = isMobile ? 5 : 7;
        const { pages, showLeftEllipsis, showRightEllipsis } = calculatePagination({
            currentPage: currPage,
            totalPages: totalPgs,
            paginationItemsToDisplay
        });

        let html = '';

        const prevDisabled = currPage <= 1;
        html += `
            <li class="pagination-item">
                <button type="button" class="pagination-btn pagination-prev" data-action="prev" ${prevDisabled ? 'disabled aria-disabled="true"' : ''} title="صفحه قبل (جدیدتر)" aria-label="صفحه قبل">
                    <svg viewBox="0 0 24 24" class="pagination-icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="m9 18 6-6-6-6"/>
                    </svg>
                </button>
            </li>
        `;

        if (showLeftEllipsis) {
            html += `
                <li class="pagination-item">
                    <button type="button" class="pagination-btn pagination-link" data-page="1" aria-label="صفحه ۱">
                        ${toPersianNum(1)}
                    </button>
                </li>
                <li class="pagination-item">
                    <span class="pagination-ellipsis" aria-hidden="true">
                        <svg viewBox="0 0 24 24" class="pagination-ellipsis-icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/><circle cx="5" cy="12" r="1"/>
                        </svg>
                        <span class="sr-only">صفحات بیشتر</span>
                    </span>
                </li>
            `;
        }

        for (const p of pages) {
            const isActive = p === currPage;
            html += `
                <li class="pagination-item">
                    <button type="button" class="pagination-btn pagination-link ${isActive ? 'is-active' : ''}" data-page="${p}" ${isActive ? 'aria-current="page"' : ''} aria-label="صفحه ${toPersianNum(p)}">
                        ${toPersianNum(p)}
                    </button>
                </li>
            `;
        }

        if (showRightEllipsis) {
            html += `
                <li class="pagination-item">
                    <span class="pagination-ellipsis" aria-hidden="true">
                        <svg viewBox="0 0 24 24" class="pagination-ellipsis-icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/><circle cx="5" cy="12" r="1"/>
                        </svg>
                        <span class="sr-only">صفحات بیشتر</span>
                    </span>
                </li>
                <li class="pagination-item">
                    <button type="button" class="pagination-btn pagination-link" data-page="${totalPgs}" aria-label="صفحه ${toPersianNum(totalPgs)}">
                        ${toPersianNum(totalPgs)}
                    </button>
                </li>
            `;
        }

        const nextDisabled = currPage >= totalPgs;
        html += `
            <li class="pagination-item">
                <button type="button" class="pagination-btn pagination-next" data-action="next" ${nextDisabled ? 'disabled aria-disabled="true"' : ''} title="صفحه بعد (قدیمی‌تر)" aria-label="صفحه بعد">
                    <svg viewBox="0 0 24 24" class="pagination-icon" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="m15 18-6-6 6-6"/>
                    </svg>
                </button>
            </li>
        `;

        controls.innerHTML = html;
    }

    function renderCurrentPage() {
        const totalItems = filteredReports.length;
        const totalPages = Math.ceil(totalItems / itemsPerPage) || 1;

        if (currentPage < 1) currentPage = 1;
        if (currentPage > totalPages) currentPage = totalPages;

        const startIndex = (currentPage - 1) * itemsPerPage;
        const endIndex = Math.min(startIndex + itemsPerPage, totalItems);
        const pageData = filteredReports.slice(startIndex, endIndex);

        renderTable(pageData);
        updateBulkToolbar();

        const infoEl = document.getElementById('pagination-info');
        if (infoEl) {
            const startFa = toPersianNum(totalItems === 0 ? 0 : startIndex + 1);
            const endFa = toPersianNum(endIndex);
            const totalFa = toPersianNum(totalItems);
            infoEl.textContent = `گزارش ${startFa} تا ${endFa} از ${totalFa}`;
        }

        renderPaginationControls(currentPage, totalPages);
    }

    function createReportStatusBadge(hasResp, isRead, isForwarded) {
        const statusBadge = document.createElement('span');
        if (hasResp) {
            statusBadge.className = 'badge badge-responded';
            statusBadge.innerHTML = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="10"></circle><path d="m9 12 2 2 4-4"></path></svg><span>پاسخ داده شده</span>`;
        } else if (isForwarded) {
            statusBadge.className = 'badge badge-in-progress';
            statusBadge.innerHTML = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M10.1 2.18a9.93 9.93 0 0 1 3.8 0"></path><path d="M17.6 3.71a9.95 9.95 0 0 1 2.69 2.7"></path><path d="M21.82 10.1a9.93 9.93 0 0 1 0 3.8"></path><path d="M20.29 17.6a9.95 9.95 0 0 1-2.7 2.69"></path><path d="M13.9 21.82a9.93 9.93 0 0 1-3.8 0"></path><path d="M6.4 20.29a9.95 9.95 0 0 1-2.69-2.7"></path><path d="M2.18 13.9a9.93 9.93 0 0 1 0-3.8"></path><path d="M3.71 6.4a9.95 9.95 0 0 1 2.7-2.69"></path></svg><span>ارجاع به منطقه</span>`;
        } else if (isRead) {
            statusBadge.className = 'badge badge-read';
            statusBadge.innerHTML = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3 7V5a2 2 0 0 1 2-2h2"></path><path d="M17 3h2a2 2 0 0 1 2 2v2"></path><path d="M21 17v2a2 2 0 0 1-2 2h-2"></path><path d="M7 21H5a2 2 0 0 1-2-2v-2"></path><circle cx="12" cy="12" r="3"></circle><path d="m16 16-1.9-1.9"></path></svg><span>بررسی شده</span>`;
        } else {
            statusBadge.className = 'badge badge-unread';
            statusBadge.innerHTML = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg><span>در نوبت بررسی</span>`;
        }
        return statusBadge;
    }

    function renderTable(reports) {
        reportsBody.innerHTML = '';

        if (reports.length === 0) {
            const tr = document.createElement('tr');
            const td = document.createElement('td');
            td.colSpan = 8;
            td.style.textAlign = 'center';
            td.style.padding = '80px 20px';
            td.style.color = '#94a3b8';
            td.textContent = 'هیچ گزارشی با فیلترهای انتخابی یافت نشد.';
            tr.appendChild(td);
            reportsBody.appendChild(tr);
            return;
        }

        if (reportsBody && !reportsBody._perfDelegated) {
            reportsBody._perfDelegated = true;

            reportsBody.addEventListener('change', (e) => {
                if (e.target.matches('.bulk-checkbox')) {
                    const cb = e.target;
                    const rId = parseInt(cb.getAttribute('data-report-id'), 10);
                    if (!isNaN(rId)) {
                        if (cb.checked) selectedReportIds.add(rId);
                        else selectedReportIds.delete(rId);
                        updateBulkToolbar();
                    }
                }
            });

            reportsBody.addEventListener('click', (e) => {
                const copyBtn = e.target.closest('button[data-copy-tracking]');
                if (copyBtn) {
                    e.stopPropagation();
                    const code = copyBtn.getAttribute('data-copy-tracking');
                    if (code) {
                        if (navigator.clipboard && navigator.clipboard.writeText) {
                            navigator.clipboard.writeText(code);
                        } else {
                            const temp = document.createElement('textarea');
                            temp.value = code;
                            document.body.appendChild(temp);
                            temp.select();
                            document.execCommand('copy');
                            document.body.removeChild(temp);
                        }
                        const origSvg = copyBtn.innerHTML;
                        copyBtn.innerHTML = '<svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="#10b981" stroke-width="2.4"><polyline points="20 6 9 17 4 12"/></svg>';
                        copyBtn.classList.add('copied');
                        showToast('کد پیگیری کپی شد', 2000, 'info');
                        setTimeout(() => {
                            copyBtn.innerHTML = origSvg;
                            copyBtn.classList.remove('copied');
                        }, 1500);
                    }
                    return;
                }

                const btn = e.target.closest('button.btn-action');
                if (!btn) return;
                e.stopPropagation();

                const tr = btn.closest('tr');
                if (!tr) return;
                const reportId = parseInt(tr.getAttribute('data-report-id'), 10);
                const trackingCode = tr.getAttribute('data-tracking-code') || '';
                if (isNaN(reportId)) return;

                if (btn.classList.contains('btn-view-report')) {
                    viewReport(reportId);
                } else if (btn.classList.contains('btn-toggle-read')) {
                    toggleReadStatus(reportId);
                } else if (btn.classList.contains('btn-forward-report')) {
                    openForwardModal(reportId, trackingCode);
                } else if (btn.classList.contains('btn-forward-history')) {
                    openForwardHistoryModal(reportId, trackingCode);
                } else if (btn.classList.contains('btn-delete-report')) {
                    deleteReportPermanently(reportId, trackingCode);
                }
            });
        }

        const fragment = document.createDocumentFragment();

        const searchTerm = activeSearchContext.rawTerm;
        const searchRegex = activeSearchContext.regex;
        const normSearch = activeSearchContext.normTerm;

        reports.forEach(report => {
            const tr = document.createElement('tr');
            tr.setAttribute('data-report-id', report.id);
            tr.setAttribute('data-tracking-code', report.tracking_code || '');
            const prio = report.priority || 'normal';
            tr.className = `row-priority-${prio}`;
            if (report.has_new_message) {
                tr.classList.add('row-new-message');
            }
            if (report.has_regional_response) {
                tr.classList.add('row-regional-response');
            }

            const checkTd = document.createElement('td');
            checkTd.className = 'bulk-td';
            const cb = document.createElement('input');
            cb.type = 'checkbox';
            cb.className = 'bulk-checkbox';
            cb.setAttribute('data-report-id', report.id);
            cb.checked = selectedReportIds.has(report.id);
            checkTd.appendChild(cb);
            tr.appendChild(checkTd);

            const shamsiDate = formatJalaliDateTime(report.created_at);

            const codeTd = document.createElement('td');
            codeTd.style.fontFamily = 'Tahoma, Arial, sans-serif';
            codeTd.style.fontWeight = 'bold';
            codeTd.style.direction = 'ltr';
            codeTd.style.textAlign = 'right';

            const trackingWrapper = document.createElement('div');
            trackingWrapper.className = 'table-tracking-wrapper';

            const codeSpan = document.createElement('span');
            codeSpan.className = 'table-tracking-code';
            codeSpan.innerHTML = highlightSearchTerm(report.tracking_code, searchTerm, searchRegex);

            const copyBtn = document.createElement('button');
            copyBtn.type = 'button';
            copyBtn.className = 'btn-table-copy';
            copyBtn.setAttribute('data-copy-tracking', report.tracking_code || '');
            copyBtn.setAttribute('data-tooltip', 'کپی کد پیگیری');
            copyBtn.setAttribute('aria-label', 'کپی کد پیگیری');
            copyBtn.innerHTML = '<svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>';

            trackingWrapper.appendChild(codeSpan);
            trackingWrapper.appendChild(copyBtn);
            codeTd.appendChild(trackingWrapper);
            tr.appendChild(codeTd);

            const priorityTd = document.createElement('td');
            const priorityBadge = document.createElement('span');
            if (prio === 'critical') {
                priorityBadge.className = 'badge-priority badge-priority-critical';
                priorityBadge.innerHTML = '<span class="priority-dot priority-dot-critical"></span>آنی';
            } else if (prio === 'urgent') {
                priorityBadge.className = 'badge-priority badge-priority-urgent';
                priorityBadge.innerHTML = '<span class="priority-dot priority-dot-urgent"></span>فوری';
            } else {
                priorityBadge.className = 'badge-priority badge-priority-normal';
                priorityBadge.innerHTML = '<span class="priority-dot priority-dot-normal"></span>عادی';
            }
            priorityTd.appendChild(priorityBadge);
            tr.appendChild(priorityTd);

            const dateTd = document.createElement('td');
            dateTd.textContent = shamsiDate;
            dateTd.style.direction = 'ltr';
            dateTd.style.textAlign = 'right';
            tr.appendChild(dateTd);

            const catTd = document.createElement('td');
            const catBadge = document.createElement('div');
            catBadge.className = 'table-category-label';
            catBadge.innerHTML = highlightSearchTerm(report.category, searchTerm, searchRegex);
            catTd.appendChild(catBadge);

            if (normSearch && report.description) {
                if ((report._normDesc && report._normDesc.includes(normSearch)) || (report._lowerDesc && report._lowerDesc.includes(searchTerm.toLowerCase()))) {
                    const snippetHtml = getMatchSnippet(report.description, searchTerm, 80, searchRegex);
                    if (snippetHtml) {
                        const snippetEl = document.createElement('div');
                        snippetEl.className = 'search-match-snippet';
                        snippetEl.innerHTML = `
                            <svg class="snippet-icon" viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                                <polyline points="14 2 14 8 20 8"></polyline>
                                <line x1="16" y1="13" x2="8" y2="13"></line>
                                <line x1="16" y1="17" x2="8" y2="17"></line>
                            </svg>
                            <span class="snippet-text">${snippetHtml}</span>
                        `;
                        catTd.appendChild(snippetEl);
                    }
                }
            }
            tr.appendChild(catTd);

            const typeTd = document.createElement('td');
            typeTd.textContent = report.report_type === 'anonymous' ? 'ناشناس' : 'با هویت';
            tr.appendChild(typeTd);

            const statusCell = document.createElement('td');
            statusCell.className = 'status-cell';
            const badgesWrapper = document.createElement('div');
            badgesWrapper.className = 'status-badges-wrap';

            let hasSpecialBadge = false;

            if (report.has_new_message) {
                const msgBadge = document.createElement('span');
                msgBadge.className = 'badge badge-new-message';
                msgBadge.setAttribute('data-tooltip', 'پیام جدید از سوی گزارش‌دهنده');
                msgBadge.innerHTML = `<span class="badge-dot-beacon"></span><svg class="badge-icon-chat" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg><span>پیام جدید</span>`;
                badgesWrapper.appendChild(msgBadge);
                hasSpecialBadge = true;
            }

            if (report.has_regional_response) {
                const regBadge = document.createElement('span');
                regBadge.className = 'badge badge-regional-response';
                regBadge.setAttribute('data-tooltip', 'پاسخ بررسی حراست از سوی ناحیه / منطقه');
                regBadge.innerHTML = `<span class="badge-dot-beacon beacon-indigo"></span><svg class="badge-icon-regional" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><polyline points="16 11 18 13 22 9"></polyline></svg><span>پاسخ حراست</span>`;
                badgesWrapper.appendChild(regBadge);
                hasSpecialBadge = true;
            }

            if (!hasSpecialBadge) {
                const hasResp = Boolean(report.has_response || (report.admin_response && report.admin_response.trim() !== ''));
                const isFwd = Boolean(report.is_forwarded);
                badgesWrapper.appendChild(createReportStatusBadge(hasResp, report.is_read, isFwd));
            }
            statusCell.appendChild(badgesWrapper);
            tr.appendChild(statusCell);

            const actionTd = document.createElement('td');
            const actionsWrap = document.createElement('div');
            actionsWrap.style.display = 'inline-flex';
            actionsWrap.style.alignItems = 'center';
            actionsWrap.style.gap = '6px';

            const viewBtn = document.createElement('button');
            viewBtn.type = 'button';
            viewBtn.className = 'btn-action btn-icon-only btn-view-report';
            viewBtn.setAttribute('data-tooltip', 'مشاهده جزئیات گزارش');
            viewBtn.setAttribute('aria-label', 'مشاهده جزئیات گزارش');
            viewBtn.innerHTML = `<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>`;
            actionsWrap.appendChild(viewBtn);

            if (!report.is_read) {
                const toggleBtn = document.createElement('button');
                toggleBtn.type = 'button';
                toggleBtn.className = 'btn-action btn-secondary btn-icon-only btn-toggle-read';
                toggleBtn.setAttribute('data-tooltip', 'علامت‌گذاری به عنوان بررسی شده');
                toggleBtn.setAttribute('aria-label', 'علامت‌گذاری به عنوان بررسی شده');
                toggleBtn.innerHTML = `<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="20 6 9 17 4 12"></polyline></svg>`;
                actionsWrap.appendChild(toggleBtn);
            }

            const canForward = currentAdmin && (currentAdmin.role === 'edare_kol' || currentAdmin.role === 'superuser' || currentAdmin.is_superuser);
            if (canForward) {
                const forwardBtn = document.createElement('button');
                forwardBtn.type = 'button';
                forwardBtn.className = 'btn-action btn-secondary btn-icon-only btn-forward-report';
                forwardBtn.setAttribute('data-tooltip', 'ارجاع خبر به کارشناسان حراست');
                forwardBtn.setAttribute('aria-label', 'ارجاع خبر به کارشناسان حراست');
                forwardBtn.innerHTML = `<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"></path><polyline points="16 6 12 2 8 6"></polyline><line x1="12" y1="2" x2="12" y2="15"></line></svg>`;
                actionsWrap.appendChild(forwardBtn);
            }

            const historyBtn = document.createElement('button');
            historyBtn.type = 'button';
            historyBtn.className = 'btn-action btn-secondary btn-icon-only btn-forward-history';
            historyBtn.setAttribute('data-tooltip', 'تاریخچه ارجاع');
            historyBtn.setAttribute('aria-label', 'تاریخچه ارجاع');
            historyBtn.innerHTML = `<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>`;
            actionsWrap.appendChild(historyBtn);

            const isSuper = Boolean(currentAdmin && (currentAdmin.is_superuser === true || currentAdmin.role === 'superuser'));
            if (isSuper) {
                const deleteBtn = document.createElement('button');
                deleteBtn.type = 'button';
                deleteBtn.className = 'btn-action btn-danger btn-icon-only btn-delete-report';
                deleteBtn.setAttribute('data-tooltip', 'حذف دائمی پرونده (پشتیبانی سامانه)');
                deleteBtn.setAttribute('aria-label', 'حذف دائمی پرونده (پشتیبانی سامانه)');
                deleteBtn.innerHTML = `<svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>`;
                actionsWrap.appendChild(deleteBtn);
            }

            actionTd.appendChild(actionsWrap);
            tr.appendChild(actionTd);
            fragment.appendChild(tr);
        });

        reportsBody.appendChild(fragment);
    }

    function closeDetailView() {
        if (detailSection) detailSection.style.display = "none";
        if (listSection) listSection.style.display = "block";
        if (filterContainer) filterContainer.style.display = "";
        reportsInboxSection?.classList.remove('in-detail-view');
        updateBulkToolbar();
        currentViewingIndex = -1;
        const deleteReportBtn = document.getElementById('btnDeleteReportSuperuser');
        if (deleteReportBtn) {
            deleteReportBtn.classList.add('d-none');
            deleteReportBtn.style.setProperty('display', 'none', 'important');
        }
        const copyPhoneBtnEl = document.getElementById('copyDetailPhoneBtn');
        if (copyPhoneBtnEl) copyPhoneBtnEl.style.display = 'none';
    }

    if (backBtn) {
        backBtn.addEventListener('click', closeDetailView);
    }

    const btnPrevReport = document.getElementById('btnPrevReportDetail');
    const btnNextReport = document.getElementById('btnNextReportDetail');

    function updateDetailNavButtons() {
        if (btnPrevReport) btnPrevReport.disabled = (currentViewingIndex <= 0);
        if (btnNextReport) btnNextReport.disabled = (currentViewingIndex < 0 || currentViewingIndex >= filteredReports.length - 1);
    }

    btnPrevReport?.addEventListener('click', () => {
        if (currentViewingIndex > 0) {
            viewReport(filteredReports[currentViewingIndex - 1].id);
        }
    });

    btnNextReport?.addEventListener('click', () => {
        if (currentViewingIndex >= 0 && currentViewingIndex < filteredReports.length - 1) {
            viewReport(filteredReports[currentViewingIndex + 1].id);
        }
    });

    const modalShortcuts = document.getElementById('modal-shortcuts');
    const btnOpenShortcuts = document.getElementById('btnOpenShortcuts');
    const btnCloseShortcuts = document.getElementById('btn-close-shortcuts-modal');
    const btnCloseShortcutsFooter = document.getElementById('btn-close-shortcuts-footer');

    function openShortcutsModal() {
        const userMenu = document.getElementById('userMenuDropdown');
        const userBtn = document.getElementById('btnUserMenu');
        if (userMenu) userMenu.style.display = 'none';
        if (userBtn) userBtn.classList.remove('is-active');
        if (modalShortcuts) {
            const card = modalShortcuts.querySelector('.modal-card, .pro-card') || modalShortcuts;

            promoteTemporarily(card, 'transform, opacity', 250);
            modalShortcuts.style.display = 'flex';
        }
    }

    function closeShortcutsModal() {
        if (modalShortcuts) modalShortcuts.style.display = 'none';
    }

    btnOpenShortcuts?.addEventListener('click', openShortcutsModal);
    btnCloseShortcuts?.addEventListener('click', closeShortcutsModal);
    btnCloseShortcutsFooter?.addEventListener('click', closeShortcutsModal);
    modalShortcuts?.addEventListener('click', (e) => {
        if (e.target === modalShortcuts) closeShortcutsModal();
    });

    document.addEventListener('keydown', (e) => {
        const isEditing = ['INPUT', 'TEXTAREA', 'SELECT'].includes(document.activeElement?.tagName);

        if ((e.metaKey || e.ctrlKey) && (e.key === 'k' || e.key === 'K')) {
            e.preventDefault();
            const reportsTab = document.getElementById('tab-btn-reports');
            if (reportsTab && !reportsTab.classList.contains('active')) {
                reportsTab.click();
            }
            const searchEl = document.getElementById('search-tracking');
            if (searchEl) {
                searchEl.focus();
                searchEl.select();
            }
            return;
        }

        if (e.key === 'Escape') {
            const actionDropdown = document.getElementById('actionSearchDropdown');
            if (actionDropdown && actionDropdown.style.display !== 'none') {
                actionDropdown.style.display = 'none';
                const searchEl = document.getElementById('search-tracking');
                if (searchEl && document.activeElement === searchEl) {
                    searchEl.blur();
                }
                return;
            }

            if (modalShortcuts && modalShortcuts.style.display !== 'none') {
                closeShortcutsModal();
                return;
            }
            if (detailSection && detailSection.style.display === 'block') {
                closeDetailView();
                return;
            }

            document.querySelectorAll('.admin-modal-overlay').forEach(m => {
                if (m.id !== 'modal-concurrent-alert') m.style.display = 'none';
            });
            return;
        }

        if (!isEditing) {
            if (e.key === '?' || (e.shiftKey && e.key === '/')) {
                e.preventDefault();
                openShortcutsModal();
            } else if (e.key === '/') {
                e.preventDefault();
                const searchEl = document.getElementById('search-tracking');
                if (searchEl) {
                    searchEl.focus();
                    searchEl.select();
                }
            } else if (detailSection && detailSection.style.display === 'block') {
                if (e.key === 'r' || e.key === 'R') {
                    e.preventDefault();
                    const respInput = document.getElementById('admin_response_input');
                    if (respInput) {
                        respInput.focus();
                        respInput.scrollIntoView({ behavior: getScrollBehavior(), block: 'center' });
                    }
                } else if (e.key === 'n' || e.key === 'N') {
                    e.preventDefault();
                    if (currentViewingIndex >= 0 && currentViewingIndex < filteredReports.length - 1) {
                        viewReport(filteredReports[currentViewingIndex + 1].id);
                    }
                } else if (e.key === 'p' || e.key === 'P') {
                    e.preventDefault();
                    if (currentViewingIndex > 0) {
                        viewReport(filteredReports[currentViewingIndex - 1].id);
                    }
                }
            }
        }
    });

    document.getElementById('btnPrintDossierBtn')?.addEventListener('click', () => {
        if (!currentViewingReportId) return;
        showToast('در حال صدور و دریافت برگ پرونده بر اساس قالب ورد...');
        window.location.href = `/api/admin/report/export-docx?id=${currentViewingReportId}`;
    });

    let _elDetailSkeleton = null;
    let _elDetailActual = null;

    async function viewReport(id, preserveScroll = false, scrollToNewMessage = false) {
        currentViewingReportId = id;
        currentViewingIndex = filteredReports.findIndex(r => r.id === id);
        updateDetailNavButtons();

        if (!_elDetailSkeleton) _elDetailSkeleton = document.getElementById('report-detail-skeleton');
        if (!_elDetailActual) _elDetailActual = document.getElementById('report-detail-actual');

        if (listSection) listSection.style.display = 'none';
        if (filterContainer) filterContainer.style.display = 'none';
        if (bulkActionsToolbar) bulkActionsToolbar.style.display = 'none';
        reportsInboxSection?.classList.add('in-detail-view');
        reportsInboxSection?.classList.remove('popover-active');
        dashboardSection?.classList.remove('popover-active');
        if (detailSection) detailSection.style.display = 'block';
        const savedScrollY = window.scrollY;
        if (!preserveScroll) {
            if (_elDetailSkeleton) _elDetailSkeleton.classList.remove('d-none');
            if (_elDetailActual) _elDetailActual.classList.add('d-none');
            window.scrollTo({ top: 0, behavior: getScrollBehavior() });
        }

        try {
            const response = await adminFetch(`/api/admin/report?id=${id}`);
            if (!response.ok) {
                if (_elDetailSkeleton) _elDetailSkeleton.classList.add('d-none');
                if (_elDetailActual) _elDetailActual.classList.remove('d-none');
                closeDetailView();
                showToast('خطا در دریافت جزئیات گزارش', 3500, 'error');
                return;
            }
            const data = await response.json();
            resetSessionTimer();

            const repInAll = allReports.find(r => r.id === id);
            if (repInAll) {
                repInAll.is_read = true;
                repInAll.has_new_message = false;
                repInAll.has_regional_response = false;
                if (!repInAll.read_by_name) {
                    repInAll.read_by_name = data.read_by_name || data.read_by || (currentAdmin ? (currentAdmin.full_name || currentAdmin.username) : '');
                }
            }
            const repInFiltered = filteredReports.find(r => r.id === id);
            if (repInFiltered) {
                repInFiltered.is_read = true;
                repInFiltered.has_new_message = false;
                repInFiltered.has_regional_response = false;
                if (!repInFiltered.read_by_name) {
                    repInFiltered.read_by_name = data.read_by_name || data.read_by || (currentAdmin ? (currentAdmin.full_name || currentAdmin.username) : '');
                }
            }

            const rowEl = reportsBody.querySelector(`tr[data-report-id="${id}"]`);
            if (rowEl) {
                rowEl.classList.remove('row-new-message');
                rowEl.classList.remove('row-regional-response');
                const cellEl = rowEl.querySelector('.status-cell');
                if (cellEl) {
                    const hasResp = (repInAll && (repInAll.has_response || (repInAll.admin_response && repInAll.admin_response.trim() !== ''))) ||
                        (data.admin_response && data.admin_response.trim() !== '');
                    cellEl.innerHTML = '';
                    const wrap = document.createElement('div');
                    wrap.className = 'status-badges-wrap';
                    wrap.appendChild(createReportStatusBadge(hasResp, true, Boolean(repInAll && repInAll.is_forwarded)));
                    cellEl.appendChild(wrap);
                }
                const toggleBtn = rowEl.querySelector('.btn-toggle-read');
                if (toggleBtn) toggleBtn.remove();
            }

            updateKPICards();
            fetchNotifications();

            const shamsiDate = formatJalaliDateTime(data.created_at, true);
            const printCurrentDate = formatJalaliDateTime(getServerNow().toISOString(), true);

            document.getElementById('detail-tracking').textContent = data.tracking_code;
            document.getElementById('detail-date').textContent = shamsiDate;

            const printTrackingEl = document.getElementById('print-header-tracking');
            const printRegDateEl = document.getElementById('print-header-regdate');
            const printDateEl = document.getElementById('print-header-date');
            if (printTrackingEl) printTrackingEl.textContent = data.tracking_code;
            if (printRegDateEl) printRegDateEl.textContent = shamsiDate;
            if (printDateEl) printDateEl.textContent = printCurrentDate;

            const prioritySelect = document.getElementById('detail-priority-select');
            const priorityPrintEl = document.getElementById('detail-priority-print');
            const priorityLabels = {
                'normal': 'عادی',
                'urgent': 'فوری',
                'critical': 'آنی و ویژه'
            };

            const updatePriorityDisplay = (prio) => {
                if (priorityPrintEl) {
                    priorityPrintEl.textContent = priorityLabels[prio] || 'عادی';
                }
            };

            if (prioritySelect) {
                const isNavahiAdmin = Boolean(currentAdmin && currentAdmin.role === 'navahi_manategh');
                const prio = data.priority || 'normal';
                prioritySelect.value = prio;
                updatePriorityDisplay(prio);

                const customPriorityDropdown = document.querySelector('.admin-custom-dropdown[data-target="detail-priority-select"]');
                let fixedBadge = document.getElementById('detail-priority-badge');
                if (isNavahiAdmin) {
                    prioritySelect.style.display = 'none';
                    if (customPriorityDropdown) customPriorityDropdown.style.display = 'none';
                    if (!fixedBadge) {
                        fixedBadge = document.createElement('span');
                        fixedBadge.id = 'detail-priority-badge';
                        prioritySelect.parentNode.insertBefore(fixedBadge, prioritySelect);
                    }
                    fixedBadge.style.display = 'inline-flex';
                    if (prio === 'critical') {
                        fixedBadge.className = 'badge-priority badge-priority-critical';
                        fixedBadge.innerHTML = '<span class="priority-dot priority-dot-critical"></span>آنی';
                    } else if (prio === 'urgent') {
                        fixedBadge.className = 'badge-priority badge-priority-urgent';
                        fixedBadge.innerHTML = '<span class="priority-dot priority-dot-urgent"></span>فوری';
                    } else {
                        fixedBadge.className = 'badge-priority badge-priority-normal';
                        fixedBadge.innerHTML = '<span class="priority-dot priority-dot-normal"></span>عادی';
                    }
                    prioritySelect.onchange = null;
                } else {
                    prioritySelect.style.display = 'none';
                    if (customPriorityDropdown) {
                        customPriorityDropdown.style.display = '';
                        const ctrl = adminCustomSelectControllers.get('detail-priority-select');
                        if (ctrl) ctrl.sync();
                    }
                    if (fixedBadge) {
                        fixedBadge.style.display = 'none';
                    }
                    prioritySelect.disabled = false;
                    prioritySelect.style.cursor = '';
                    prioritySelect.style.opacity = '';
                    prioritySelect.title = '';
                    prioritySelect.onchange = async () => {
                        const newPrio = prioritySelect.value;
                        updatePriorityDisplay(newPrio);
                        try {
                            const res = await adminFetch('/api/admin/report/set-priority', {
                                method: 'POST',
                                headers: { 'Content-Type': 'application/json' },
                                body: JSON.stringify({ report_id: id, priority: newPrio })
                            });
                            if (res.ok) {
                                showToast('اولویت پرونده با موفقیت تغییر یافت', 2500, 'success');
                                fetchReports().catch(() => { });
                            } else {
                                const errText = await res.text();
                                showToast(errText || 'خطا در تغییر اولویت', 3000, 'error');
                            }
                        } catch (err) {
                            showToast('خطا در ارتباط با سرور', 3000, 'error');
                        }
                    };
                }
            }

            const readAtTd = document.getElementById('detail-read-at');
            const adminName = data.read_by_name || data.read_by || (currentAdmin ? (currentAdmin.full_name || currentAdmin.username) : '');
            const readTimestamp = data.read_at || getServerNow().toISOString();
            const shamsiReadDate = formatJalaliDateTime(readTimestamp, true);
            const byText = adminName ? ` توسط ${adminName}` : '';

            readAtTd.textContent = `مشاهده شده در ${shamsiReadDate}${byText}`;
            readAtTd.className = 'detail-value rtl text-success';
            readAtTd.style.direction = 'rtl';

            document.getElementById('detail-category').textContent = data.category;
            document.getElementById('detail-type').textContent = data.report_type === 'anonymous' ? 'ناشناس' : 'با هویت';
            document.getElementById('detail-name').textContent = data.reporter_name || 'ثبت نشده';
            const phoneVal = (data.reporter_phone || '').trim();
            const hasPhone = Boolean(phoneVal && phoneVal !== 'ثبت نشده');
            document.getElementById('detail-phone').textContent = hasPhone ? phoneVal : 'ثبت نشده';
            const copyPhoneBtnEl = document.getElementById('copyDetailPhoneBtn');
            if (copyPhoneBtnEl) {
                copyPhoneBtnEl.style.display = hasPhone ? 'inline-flex' : 'none';
            }
            const detailDescEl = document.getElementById('detail-desc');
            if (detailDescEl) {
                const activeSearch = searchTracking ? searchTracking.value.trim() : '';
                if (activeSearch) {
                    detailDescEl.innerHTML = highlightSearchTerm(data.description, activeSearch);
                } else {
                    detailDescEl.textContent = data.description;
                }
            }

            updateClassificationUI(data, id);

            const messagesContainer = document.getElementById('detail-messages-container');
            const messagesList = document.getElementById('detail-messages-list');
            const messagesCountBadge = document.getElementById('detail-messages-count-badge');
            if (messagesContainer && messagesList) {
                messagesList.innerHTML = '';
                messagesContainer.classList.remove('d-none');

                const allItems = [];

                if (Array.isArray(data.messages) && data.messages.length > 0) {
                    let hasAdminInMessages = false;
                    data.messages.forEach(msg => {
                        const isReporter = msg.sender_type === 'reporter';
                        if (!isReporter) hasAdminInMessages = true;
                        const adminDisplayName = msg.admin_full_name || msg.admin_username || '';
                        const adminName = adminDisplayName ? ` (${escapeHTML(adminDisplayName)})` : '';
                        allItems.push({
                            sender_type: msg.sender_type,
                            is_official_response: !isReporter,
                            sender_title: isReporter ? 'گزارش‌دهنده' : ('پاسخ اداره حراست' + adminName),
                            message: msg.message,
                            created_at: msg.created_at
                        });
                    });
                    if (!hasAdminInMessages && data.admin_response && data.admin_response.trim() !== '') {
                        const respondedByTitle = data.responded_by ? ` (ثبت توسط ${escapeHTML(data.responded_by)})` : '';
                        allItems.push({
                            sender_type: 'admin',
                            is_official_response: true,
                            sender_title: 'پاسخ و نتیجه بررسی اداره حراست' + respondedByTitle,
                            message: data.admin_response.trim(),
                            created_at: data.responded_at || data.created_at
                        });
                    }
                } else if (data.admin_response && data.admin_response.trim() !== '') {
                    const respondedByTitle = data.responded_by ? ` (ثبت توسط ${escapeHTML(data.responded_by)})` : '';
                    allItems.push({
                        sender_type: 'admin',
                        is_official_response: true,
                        sender_title: 'پاسخ و نتیجه بررسی اداره حراست' + respondedByTitle,
                        message: data.admin_response.trim(),
                        created_at: data.responded_at || data.created_at
                    });
                }

                allItems.sort((a, b) => new Date(a.created_at || 0) - new Date(b.created_at || 0));

                if (messagesCountBadge) {
                    messagesCountBadge.textContent = allItems.length > 0 ? `${toPersianNum(allItems.length)} مورد ثبت‌شده` : 'فاقد پاسخ یا پیام تکمیلی';
                }

                if (allItems.length > 0) {
                    allItems.forEach(item => {
                        const isReporter = item.sender_type === 'reporter';
                        const row = document.createElement('div');
                        row.className = `nexus-bubble-row ${isReporter ? 'nexus-row-reporter' : 'nexus-row-admin'}`;

                        const userSvg = '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>';
                        const shieldSvg = '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>';

                        const avatarSvg = isReporter ? userSvg : shieldSvg;
                        const avatarClass = isReporter ? 'nexus-avatar-reporter' : 'nexus-avatar-admin';
                        const bubbleCardClass = isReporter ? 'nexus-bubble-reporter' : (item.is_official_response ? 'nexus-bubble-admin nexus-bubble-official' : 'nexus-bubble-admin');

                        const timeStr = item.created_at ? formatJalaliDateTime(item.created_at) : '';

                        row.innerHTML = `
                            <div class="nexus-bubble-unit">
                                <div class="nexus-bubble-avatar ${avatarClass}" title="${escapeHTML(item.sender_title)}">
                                    ${avatarSvg}
                                </div>
                                <div class="nexus-bubble-card ${bubbleCardClass}">
                                    <div class="nexus-bubble-top">
                                        <span class="nexus-bubble-sender">${escapeHTML(item.sender_title)}</span>
                                        ${timeStr ? `<span class="nexus-bubble-timestamp">${timeStr}</span>` : ''}
                                    </div>
                                    <div class="nexus-bubble-body">${escapeHTML(item.message).replace(/\n/g, '<br>')}</div>
                                </div>
                            </div>
                        `;
                        messagesList.appendChild(row);
                    });
                } else {
                    const emptyCard = document.createElement('div');
                    emptyCard.className = 'nexus-chat-empty';
                    emptyCard.innerHTML = `
                        <div class="nexus-chat-empty-icon">
                            <svg viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                            </svg>
                        </div>
                        <span class="nexus-chat-empty-title">هنوز گفتگو یا پاسخی در این پرونده ثبت نشده است</span>
                        <span class="nexus-chat-empty-sub">پیام‌های تکمیلی واصله از گزارش‌دهنده و پاسخ‌های رسمی صادرشده توسط حراست در این بخش به صورت یکپارچه و به ترتیب زمانی نمایش داده می‌شوند.</span>
                    `;
                    messagesList.appendChild(emptyCard);
                }
            }

            const notesInput = document.getElementById('detail-internal-notes-input');
            const printNotesText = document.getElementById('print-internal-notes-text');
            const saveNotesBtn = document.getElementById('btnSaveInternalNotes');
            const notesStatus = document.getElementById('internalNotesStatusMsg');

            const internalNotesVal = data.internal_notes || '';
            if (notesInput) notesInput.value = internalNotesVal;
            if (printNotesText) printNotesText.textContent = internalNotesVal || 'فاقد یادداشت محرمانه داخلی.';
            if (notesStatus) notesStatus.style.display = 'none';

            const isNavahi = currentAdmin && currentAdmin.role === 'navahi_manategh';

            if (saveNotesBtn) {
                if (isNavahi) {
                    if (notesInput) {
                        notesInput.readOnly = true;
                        notesInput.placeholder = 'فاقد یادداشت محرمانه ثبت‌شده توسط اداره کل.';
                    }
                    saveNotesBtn.style.display = 'none';
                } else {
                    if (notesInput) {
                        notesInput.readOnly = false;
                        notesInput.placeholder = 'ثبت اقدامات داخلی، شماره نامه‌های اداری، ارجاعات و یادداشت‌های محرمانه کارشناس...';
                    }
                    saveNotesBtn.style.display = '';
                    saveNotesBtn.disabled = false;
                    saveNotesBtn.style.opacity = '';
                    saveNotesBtn.style.cursor = '';
                    saveNotesBtn.onclick = async () => {
                        try {
                            const notesVal = notesInput ? notesInput.value.trim() : '';
                            const res = await adminFetch('/api/admin/report/internal-notes', {
                                method: 'POST',
                                headers: { 'Content-Type': 'application/json' },
                                body: JSON.stringify({ report_id: id, internal_notes: notesVal, notes: notesVal })
                            });
                            if (res.ok) {
                                showToast('یادداشت محرمانه ذخیره و رمزنگاری شد');
                                if (printNotesText) printNotesText.textContent = notesVal || 'فاقد یادداشت محرمانه داخلی.';
                                if (notesStatus) {
                                    notesStatus.className = 'response-box badge-read';
                                    notesStatus.textContent = 'یادداشت محرمانه با موفقیت در بانک اطلاعاتی رمزنگاری و ثبت گردید.';
                                    notesStatus.style.display = 'block';
                                }
                            } else {
                                const err = await res.text();
                                showToast(err || 'خطا در ذخیره یادداشت', 3500, 'error');
                            }
                        } catch (err) {
                            showToast('خطای ارتباط با سرور', 3500, 'error');
                        }
                    };
                }
            }

            const attachmentTd = document.getElementById('detail-attachment');
            attachmentTd.innerHTML = '';

            const formatAdminFileSize = (bytes) => {
                if (!bytes || bytes === 0) return '';
                const k = 1024;
                const sizes = ['بایت', 'کیلوبایت', 'مگابایت'];
                const i = Math.floor(Math.log(bytes) / Math.log(k));
                const formatted = parseFloat((bytes / Math.pow(k, i)).toFixed(1));
                return `${toPersianNum(formatted)} ${sizes[i]}`;
            };

            if (data.attachments && data.attachments.length > 0) {
                if (data.attachments.length > 1) {
                    const bulkBtn = document.createElement('button');
                    bulkBtn.className = 'btn-attachment-bulk';
                    bulkBtn.setAttribute('data-tooltip', 'دانلود یکجای تمامی مدارک پیوست به صورت فایل فشرده');
                    bulkBtn.innerHTML = `
                        <span class="attachment-icon-box">
                            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                                <polyline points="7 10 12 15 17 10"></polyline>
                                <line x1="12" y1="15" x2="12" y2="3"></line>
                            </svg>
                        </span>
                        <span class="attachment-text-group">
                            <span class="attachment-primary-title">دانلود پیوست‌های گزارش</span>
                            <span class="attachment-sub-meta">${toPersianNum(data.attachments.length)} فایل ضمیمه • قالب فشرده ZIP</span>
                        </span>
                    `;
                    bulkBtn.onclick = () => {
                        window.location.href = `/api/admin/download?id=${id}`;
                    };
                    attachmentTd.appendChild(bulkBtn);
                } else {
                    const att = data.attachments[0];
                    const downloadBtn = document.createElement('button');
                    downloadBtn.className = 'btn-attachment-bulk';
                    const szText = att.file_size ? ` (${formatAdminFileSize(att.file_size)})` : '';
                    const nameText = att.file_name || 'فایل ضمیمه';
                    downloadBtn.setAttribute('data-tooltip', `دانلود مستقیم: ${nameText}`);

                    downloadBtn.innerHTML = `
                        <span class="attachment-icon-box">
                            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                                <polyline points="7 10 12 15 17 10"></polyline>
                                <line x1="12" y1="15" x2="12" y2="3"></line>
                            </svg>
                        </span>
                        <span class="attachment-text-group">
                            <span class="attachment-primary-title">دانلود فایل پیوست</span>
                            <span class="attachment-sub-meta">${escapeHTML(nameText)}${szText}</span>
                        </span>
                    `;
                    downloadBtn.onclick = () => {
                        window.location.href = `/api/admin/download?id=${id}&file_id=${att.id}`;
                    };
                    attachmentTd.appendChild(downloadBtn);
                }
            } else if (data.has_attachment) {
                const downloadBtn = document.createElement('button');
                downloadBtn.className = 'btn-attachment-bulk';
                downloadBtn.innerHTML = `
                    <span class="attachment-icon-box">
                        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                            <polyline points="7 10 12 15 17 10"></polyline>
                            <line x1="12" y1="15" x2="12" y2="3"></line>
                        </svg>
                    </span>
                    <span class="attachment-text-group">
                        <span class="attachment-primary-title">دانلود فایل پیوست</span>
                        <span class="attachment-sub-meta">فایل ضمیمه گزارش</span>
                    </span>
                `;
                downloadBtn.onclick = () => {
                    window.location.href = `/api/admin/download?id=${id}`;
                };
                attachmentTd.appendChild(downloadBtn);
            } else {
                attachmentTd.textContent = 'ندارد';
            }

            const existingBox = document.getElementById('existing-response-box');
            const responseTextEl = document.getElementById('detail-admin-response-text');
            const metaInfoEl = document.getElementById('response-meta-info');
            const responseInput = document.getElementById('admin_response_input');
            const statusMsg = document.getElementById('responseStatusMsg');
            const printNoRespEl = document.getElementById('print-no-response-placeholder');

            if (statusMsg) statusMsg.style.display = 'none';

            if (data.admin_response && data.admin_response.trim() !== '') {
                responseTextEl.textContent = data.admin_response;
                if (responseInput) responseInput.value = '';

                let respDateStr = '';
                if (data.responded_at) {
                    respDateStr = formatJalaliDateTime(data.responded_at, false);
                }

                metaInfoEl.textContent = `ثبت پاسخ توسط: ${data.responded_by || 'حراست'}  |  تاریخ ثبت پاسخ: ${respDateStr || 'ثبت نشده'}`;
                existingBox.style.display = 'none';
                if (printNoRespEl) printNoRespEl.style.display = 'none';
            } else {
                existingBox.style.display = 'none';
                responseTextEl.textContent = '';
                if (responseInput) responseInput.value = '';
                if (printNoRespEl) printNoRespEl.style.display = 'block';
            }

            const adminCannedWrapper = document.getElementById('admin-canned-wrapper');
            const regionalCannedWrapper = document.getElementById('regional-canned-wrapper');
            const submitRespBtn = document.getElementById('submitResponseBtn');
            const adminRespForm = document.getElementById('adminResponseForm');
            const responseHeading = document.getElementById('detail-response-heading');
            const detailRespContainer = document.querySelector('.detail-response-container');

            if (isNavahi) {
                if (adminRespForm) adminRespForm.style.display = 'none';
                if (adminCannedWrapper) adminCannedWrapper.style.display = 'none';
                if (regionalCannedWrapper) regionalCannedWrapper.style.display = '';

                const hasFinalResponse = Boolean(data.admin_response && data.admin_response.trim() !== '');
                if (hasFinalResponse) {
                    if (detailRespContainer) detailRespContainer.style.display = '';
                    if (responseHeading) responseHeading.textContent = '۲. پاسخ نهایی اداره کل حراست به گزارش‌دهنده:';
                    if (existingBox) existingBox.style.display = 'none';
                } else {
                    if (detailRespContainer) detailRespContainer.style.display = 'none';
                }
            } else {
                if (detailRespContainer) detailRespContainer.style.display = '';
                if (responseHeading) responseHeading.textContent = '۲. ارسال پاسخ رسمی به گزارش‌دهنده:';
                if (adminRespForm) adminRespForm.style.display = '';
                if (adminCannedWrapper) adminCannedWrapper.style.display = '';
                if (regionalCannedWrapper) regionalCannedWrapper.style.display = 'none';
                if (responseInput) {
                    responseInput.readOnly = false;
                    responseInput.placeholder = 'متن پاسخ یا نتیجه بررسی حراست را جهت ارسال به گزارش‌دهنده وارد نمایید...';
                }
                if (submitRespBtn) {
                    submitRespBtn.disabled = false;
                    submitRespBtn.style.opacity = '';
                    submitRespBtn.style.cursor = '';
                }
            }

            const regionalSection = document.getElementById('detail-regional-responses-container');
            const regionalCentralView = document.getElementById('regional-responses-central-view');
            const regionalSideView = document.getElementById('regional-responses-side-view');
            const regionalList = document.getElementById('regional-responses-list');

            if (regionalSection) {
                regionalSection.style.display = 'block';

                if (!isNavahi) {

                    if (regionalCentralView) regionalCentralView.style.display = 'block';
                    if (regionalSideView) regionalSideView.style.display = 'none';

                    if (regionalList) {
                        regionalList.innerHTML = '';
                        if (data.regional_responses && data.regional_responses.length > 0) {
                            data.regional_responses.forEach(rr => {
                                const card = document.createElement('div');
                                card.className = 'regional-response-card';
                                const timeFa = rr.responded_at ? formatJalaliDateTime(rr.responded_at, true) : 'ثبت‌شده';
                                card.innerHTML = `
                                    <div class="regional-response-card-header">
                                        <div class="regional-card-author">
                                            <span class="badge-regional font-bold">${escapeHTML(rr.recipient_title || 'حراست ناحیه/منطقه')}</span>
                                            <span class="regional-card-name">${escapeHTML(rr.recipient_name || rr.recipient_username)}</span>
                                        </div>
                                        <div class="regional-card-date font-tahoma ltr">
                                            <span>تاریخ ارسال:</span>
                                            <strong>${timeFa}</strong>
                                        </div>
                                    </div>
                                    <div class="regional-card-body">${escapeHTML(rr.response_text)}</div>
                                `;
                                regionalList.appendChild(card);
                            });
                        } else {
                            regionalList.innerHTML = `
                                <div class="regional-responses-empty">
                                    <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-bottom: 6px; opacity: 0.65;">
                                        <circle cx="12" cy="12" r="10"></circle>
                                        <polyline points="12 6 12 12 16 14"></polyline>
                                    </svg>
                                    <div>هنوز پاسخی از سوی کارشناسان نواحی و مناطق برای این پرونده ثبت نگردیده است.</div>
                                </div>
                            `;
                        }
                    }
                } else {

                    if (regionalCentralView) regionalCentralView.style.display = 'none';
                    if (regionalSideView) regionalSideView.style.display = 'block';

                    const prevCard = document.getElementById('regional-previous-response-card');
                    const prevTime = document.getElementById('regional-previous-time');
                    const prevText = document.getElementById('regional-previous-text');
                    const regionalInput = document.getElementById('regional_response_input');
                    const regionalStatus = document.getElementById('regionalResponseStatusMsg');

                    if (regionalStatus) regionalStatus.style.display = 'none';

                    if (data.my_regional_response && data.my_regional_response.response_text) {
                        if (prevCard) prevCard.style.display = 'block';
                        if (prevTime) prevTime.textContent = data.my_regional_response.responded_at ? formatJalaliDateTime(data.my_regional_response.responded_at, true) : '';
                        if (prevText) prevText.textContent = data.my_regional_response.response_text;
                        if (regionalInput) regionalInput.value = data.my_regional_response.response_text;
                    } else {
                        if (prevCard) prevCard.style.display = 'none';
                        if (regionalInput) regionalInput.value = '';
                    }
                }
            }

            const deleteReportBtn = document.getElementById('btnDeleteReportSuperuser');
            if (deleteReportBtn) {
                const isSuper = Boolean(currentAdmin && (currentAdmin.is_superuser === true || currentAdmin.role === 'superuser'));
                if (isSuper) {
                    deleteReportBtn.classList.remove('d-none');
                    deleteReportBtn.style.setProperty('display', 'inline-flex', 'important');
                    deleteReportBtn.onclick = () => {
                        deleteReportPermanently(id, data.tracking_code, true);
                    };
                } else {
                    deleteReportBtn.classList.add('d-none');
                    deleteReportBtn.style.setProperty('display', 'none', 'important');
                    deleteReportBtn.onclick = null;
                }
            }

            const btnForwardDetail = document.getElementById('btnForwardReportDetail');
            const canForward = currentAdmin && (currentAdmin.role === 'edare_kol' || currentAdmin.role === 'superuser' || currentAdmin.is_superuser);
            if (btnForwardDetail) {
                if (canForward) {
                    btnForwardDetail.classList.remove('d-none');
                    btnForwardDetail.onclick = () => openForwardModal(id, data.tracking_code);
                } else {
                    btnForwardDetail.classList.add('d-none');
                }
            }

            const btnHistoryDetail = document.getElementById('btnHistoryReportDetail');
            if (btnHistoryDetail) {
                btnHistoryDetail.onclick = () => openForwardHistoryModal(id, data.tracking_code);
            }

            listSection.style.display = "none";
            if (filterContainer) filterContainer.style.display = 'none';
            if (bulkActionsToolbar) bulkActionsToolbar.style.display = 'none';
            reportsInboxSection?.classList.add('in-detail-view');
            if (_elDetailSkeleton) _elDetailSkeleton.classList.add('d-none');
            if (_elDetailActual) _elDetailActual.classList.remove('d-none');
            if (window.TatweelTypography) window.TatweelTypography.recheckOverflow();

            if (notesInput) makeTextareaAutoExpand(notesInput);
            if (responseInput) makeTextareaAutoExpand(responseInput);
            const regionalTextarea = document.getElementById('regional_response_input');
            if (regionalTextarea) makeTextareaAutoExpand(regionalTextarea);

            if (!data.is_read) {
                toggleReadStatus(id, false);
            }

            if (scrollToNewMessage) {
                const msgList = document.getElementById('detail-messages-list');
                const lastBubble = msgList ? msgList.querySelector('.nexus-bubble-row:last-child') : null;
                if (lastBubble) {
                    if (msgList) msgList.scrollTop = msgList.scrollHeight;
                    requestAnimationFrame(() => {
                        lastBubble.scrollIntoView({ behavior: getScrollBehavior(), block: 'start' });
                        const bubbleCard = lastBubble.querySelector('.nexus-bubble-card');
                        if (bubbleCard) {
                            bubbleCard.classList.add('nexus-bubble-highlight');
                            setTimeout(() => bubbleCard.classList.remove('nexus-bubble-highlight'), 2200);
                        }
                    });
                } else if (preserveScroll) {
                    window.scrollTo({ top: savedScrollY, behavior: 'instant' });
                }
            } else if (preserveScroll) {
                window.scrollTo({ top: savedScrollY, behavior: 'instant' });
            }
        } catch (error) {
            console.error('Error opening report modal:', error);
            if (_elDetailSkeleton) _elDetailSkeleton.classList.add('d-none');
            if (_elDetailActual) _elDetailActual.classList.remove('d-none');
            if (!preserveScroll) {
                closeDetailView();
            }
            showToast('خطا در بارگذاری اطلاعات پرونده', 3500, 'error');
        }
    }

    const regionalResponseForm = document.getElementById('regionalResponseForm');
    regionalResponseForm?.addEventListener('submit', async (e) => {
        e.preventDefault();
        if (!currentViewingReportId) return;

        const regionalInput = document.getElementById('regional_response_input');
        const submitBtn = document.getElementById('btnSubmitRegionalResponse');
        const statusMsg = document.getElementById('regionalResponseStatusMsg');

        const text = regionalInput ? regionalInput.value.trim() : '';
        if (!text) {
            showToast('لطفاً متن گزارش و نتیجه بررسی را وارد فرمایید.', 3500, 'error');
            return;
        }

        const btnText = submitBtn?.querySelector('.btn-text');
        const spinner = submitBtn?.querySelector('.btn-spinner');
        if (submitBtn) submitBtn.disabled = true;
        if (btnText) btnText.textContent = 'در حال ارسال...';
        if (spinner) spinner.classList.remove('d-none');
        if (statusMsg) statusMsg.style.display = 'none';

        try {
            const res = await adminFetch('/api/admin/report/regional-response', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ report_id: currentViewingReportId, response: text })
            });

            if (!res.ok) {
                const err = await res.text();
                throw new Error(err || 'خطا در ثبت و ارسال گزارش');
            }

            const data = await res.json();
            showToast(data.message || 'گزارش بررسی با موفقیت برای اداره کل ارسال گردید.');
            if (statusMsg) {
                statusMsg.className = 'response-box badge-read';
                statusMsg.textContent = data.message || 'گزارش بررسی با موفقیت برای اداره کل ارسال شد.';
                statusMsg.style.display = 'block';
            }
            viewReport(currentViewingReportId, true);
        } catch (err) {
            showToast(err.message || 'خطای ارتباط با سرور', 4000, 'error');
            if (statusMsg) {
                statusMsg.className = 'response-box badge-unread';
                statusMsg.textContent = err.message || 'خطای ارتباط با سرور.';
                statusMsg.style.display = 'block';
            }
        } finally {
            if (submitBtn) submitBtn.disabled = false;
            if (btnText) btnText.textContent = 'ارسال گزارش بررسی به اداره کل';
            if (spinner) spinner.classList.add('d-none');
        }
    });

    async function deleteReportPermanently(reportId, trackingCode, isInsideDetail = false) {
        if (!currentAdmin || !currentAdmin.is_superuser) {
            showToast('تنها پشتیبانی سامانه مجاز به حذف پرونده می‌باشد.', 3000, 'warning');
            return;
        }

        const confirmed = await showConfirmDialog({
            title: 'حذف دائمی و غیرقابل بازگشت پرونده',
            description: `اخطار امنیتی پشتیبانی سامانه: آیا از حذف دائمی پرونده [${trackingCode || reportId}] و کلیه فایل‌های ضمیمه آن از سرور و پایگاه داده اطمینان دارید؟ این عملیات غیرقابل بازگشت است و در لاگ‌های بازرسی ثبت می‌گردد.`,
            confirmText: 'بله، حذف دائمی پرونده',
            cancelText: 'انصراف',
            variant: 'danger',
            iconType: 'trash'
        });
        if (!confirmed) return;

        try {
            const response = await adminFetch('/api/admin/report/delete', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ report_id: reportId })
            });

            if (!response.ok) {
                const errText = await response.text();
                throw new Error(errText || 'خطا در حذف پرونده');
            }

            const data = await response.json();
            showToast(data.message || 'پرونده با موفقیت حذف گردید', 'success');

            if (isInsideDetail && backBtn) {
                backBtn.click();
            }
            await fetchReports().catch(() => { });
        } catch (err) {
            showToast('خطا در اجرای عملیات حذف: ' + err.message, 4000, 'error');
        }
    }

    async function toggleReadStatus(id, shouldRefresh = true) {
        try {
            const payload = { report_id: id };
            const response = await adminFetch('/api/admin/toggle-read', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (response.ok) {
                resetSessionTimer();

                const repInAll = allReports.find(r => r.id === id);
                if (repInAll) {
                    repInAll.is_read = true;
                    repInAll.has_new_message = false;
                    repInAll.has_regional_response = false;
                    if (!repInAll.read_by_name && currentAdmin) {
                        repInAll.read_by_name = currentAdmin.full_name || currentAdmin.username;
                    }
                }
                const repInFiltered = filteredReports.find(r => r.id === id);
                if (repInFiltered) {
                    repInFiltered.is_read = true;
                    repInFiltered.has_new_message = false;
                    repInFiltered.has_regional_response = false;
                    if (!repInFiltered.read_by_name && currentAdmin) {
                        repInFiltered.read_by_name = currentAdmin.full_name || currentAdmin.username;
                    }
                }

                const rowEl = reportsBody.querySelector(`tr[data-report-id="${id}"]`);
                if (rowEl) {
                    rowEl.classList.remove('row-new-message');
                    rowEl.classList.remove('row-regional-response');
                    const cellEl = rowEl.querySelector('.status-cell');
                    if (cellEl) {
                        const hasResp = repInAll && (repInAll.has_response || (repInAll.admin_response && repInAll.admin_response.trim() !== ''));
                        cellEl.innerHTML = '';
                        const wrap = document.createElement('div');
                        wrap.className = 'status-badges-wrap';
                        wrap.appendChild(createReportStatusBadge(hasResp, true, Boolean(repInAll && repInAll.is_forwarded)));
                        cellEl.appendChild(wrap);
                    }
                    const toggleBtn = rowEl.querySelector('.btn-toggle-read');
                    if (toggleBtn) toggleBtn.remove();
                }

                updateKPICards();
                fetchNotifications();

                if (shouldRefresh) fetchReports().catch(() => { });
            }
        } catch (error) {
            console.error('Failed to toggle read status');
        }
    }

    function makeTextareaAutoExpand(el) {
        if (!el) return;
        el.style.resize = 'none';
        el.style.overflowY = 'hidden';
        const minHeight = el.id === 'admin_response_input' ? 85 : 70;
        let rafId;
        const adjustHeight = () => {
            cancelAnimationFrame(rafId);
            rafId = requestAnimationFrame(() => {
                el.style.height = 'auto';
                el.style.height = Math.max(el.scrollHeight, minHeight) + 'px';
            });
        };
        el.addEventListener('input', adjustHeight);
        el.addEventListener('change', adjustHeight);
        el.addEventListener('paste', () => setTimeout(adjustHeight, 0));
        adjustHeight();
        return adjustHeight;
    }

    document.querySelectorAll('textarea').forEach(ta => makeTextareaAutoExpand(ta));

    document.querySelectorAll('.canned-chip').forEach(chip => {
        chip.addEventListener('click', () => {
            const text = chip.getAttribute('data-text');
            const targetId = chip.getAttribute('data-target') || 'admin_response_input';
            const targetInput = document.getElementById(targetId);
            if (targetInput && text) {
                targetInput.value = text;
                makeTextareaAutoExpand(targetInput);
                targetInput.focus();
            }
        });
    });

    const copyTrackingBtn = document.getElementById('copyDetailTrackingBtn');
    const copyPhoneBtn = document.getElementById('copyDetailPhoneBtn');

    if (copyTrackingBtn) {
        copyTrackingBtn.addEventListener('click', () => {
            const text = document.getElementById('detail-tracking')?.textContent.trim();
            if (text) {
                navigator.clipboard?.writeText(text);
                showToast('کد پیگیری کپی شد');
            }
        });
    }

    if (copyPhoneBtn) {
        copyPhoneBtn.addEventListener('click', () => {
            const text = document.getElementById('detail-phone')?.textContent.trim();
            if (text && text !== 'ثبت نشده') {
                navigator.clipboard?.writeText(text);
                showToast('شماره تماس کپی شد');
            }
        });
    }

    const adminResponseForm = document.getElementById('adminResponseForm');
    if (adminResponseForm) {
        adminResponseForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (!currentViewingReportId) return;

            const responseInput = document.getElementById('admin_response_input');
            const statusMsg = document.getElementById('responseStatusMsg');

            if (currentAdmin && currentAdmin.role === 'navahi_manategh') {
                showToast('شما به این قسمت دسترسی ندارید.', 3500, 'error');
                if (statusMsg) {
                    statusMsg.className = 'response-box badge-unread';
                    statusMsg.textContent = 'شما به این قسمت دسترسی ندارید.';
                    statusMsg.style.display = 'block';
                }
                return;
            }
            const submitBtn = document.getElementById('submitResponseBtn');
            const btnText = submitBtn?.querySelector('.btn-text');
            const btnSpinner = submitBtn?.querySelector('.btn-spinner');

            if (submitBtn) submitBtn.disabled = true;
            if (btnText) btnText.textContent = 'در حال ارسال پاسخ...';
            if (btnSpinner) btnSpinner.classList.remove('d-none');

            try {
                const payload = {
                    report_id: currentViewingReportId,
                    response: responseInput.value.trim()
                };

                const res = await adminFetch('/api/admin/respond', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });

                if (res.ok) {
                    resetSessionTimer();
                    showToast('پاسخ با موفقیت ارسال و رمزنگاری شد');
                    if (statusMsg) {
                        statusMsg.className = 'response-box badge-read';
                        statusMsg.textContent = 'پاسخ با موفقیت ارسال و رمزنگاری شد.';
                        statusMsg.style.display = 'block';
                    }
                    if (responseInput) responseInput.value = '';
                    await viewReport(currentViewingReportId, true, true);
                    fetchReports().catch(() => { });
                } else {
                    const err = await res.text();
                    if (statusMsg) {
                        statusMsg.className = 'response-box badge-unread';
                        statusMsg.textContent = `خطا در ارسال پاسخ: ${err}`;
                        statusMsg.style.display = 'block';
                    }
                }
            } catch (error) {
                if (statusMsg) {
                    statusMsg.className = 'response-box badge-unread';
                    statusMsg.textContent = 'خطای ارتباط با سرور.';
                    statusMsg.style.display = 'block';
                }
            } finally {
                if (submitBtn) submitBtn.disabled = false;
                if (btnText) btnText.textContent = 'ارسال پاسخ';
                if (btnSpinner) btnSpinner.classList.add('d-none');
            }
        });
    }

    async function handleAdminLogout() {
        const confirmed = await showConfirmDialog({
            title: 'خروج از کارتابل حراست',
            description: 'آیا از خروج از حساب کاربری اداری خود اطمینان دارید؟ برای دسترسی مجدد نیاز به ورود با نام کاربری و گذرواژه خواهید داشت.',
            confirmText: 'بله، خروج از سامانه',
            cancelText: 'انصراف و ماندن',
            variant: 'warning',
            iconType: 'logout'
        });
        if (!confirmed) return;

        try {
            await adminFetch('/api/admin/logout', { method: 'POST' });
        } catch (err) {
            console.error('Logout error:', err);
        } finally {
            sessionStorage.removeItem('2fa_prompt_shown');
            document.cookie = 'herasat_admin_session=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
            location.reload();
        }
    }

    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) logoutBtn.addEventListener('click', handleAdminLogout);

    if (btnLogoutSessionWarning) btnLogoutSessionWarning.addEventListener('click', handleAdminLogout);

    function updateClassificationUI(data, reportId) {
        const bannerEl = document.getElementById('detail-classification-note');
        const noteTextEl = document.getElementById('classificationNoteText');
        const noteIconEl = document.getElementById('classificationNoteIcon');
        const actionContainer = document.getElementById('classificationActionContainer');
        const promptContainer = document.getElementById('classification-prompt-container');

        const isNavahi = currentAdmin && currentAdmin.role === 'navahi_manategh';

        const classification = data ? data.classification : null;
        if (classification && (classification === 'helpful' || classification === 'useless')) {

            if (promptContainer) promptContainer.style.display = 'none';

            if (bannerEl && noteTextEl && noteIconEl) {
                bannerEl.style.display = 'block';
                bannerEl.className = `detail-classification-banner ${classification === 'helpful' ? 'banner-helpful' : 'banner-useless'}`;

                const classifierName = data.classified_by_name || data.classified_by || (currentAdmin ? (currentAdmin.full_name || currentAdmin.username) : 'کارشناس حراست');

                if (classification === 'helpful') {
                    noteIconEl.innerHTML = `<svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M14 9V5a3 3 0 0 0-3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0-2-2.3zM7 22H4a2 2 0 0 1-2-2v-7a2 2 0 0 1 2-2h3"/>
                    </svg>`;
                    noteTextEl.innerHTML = `این گزارش توسط <strong>${escapeHTML(classifierName)}</strong> به عنوان <strong class="badge-tag-helpful">مفید و موثر</strong> علامت‌گذاری شده است.`;
                } else {
                    noteIconEl.innerHTML = `<svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10"/>
                        <line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/>
                    </svg>`;
                    noteTextEl.innerHTML = `این گزارش توسط <strong>${escapeHTML(classifierName)}</strong> به عنوان <strong class="badge-tag-useless">بی‌محتوا و فاقد ارزش</strong> علامت‌گذاری شده است.`;
                }

                const isClassifier = !isNavahi && currentAdmin && data.classified_by && (currentAdmin.username === data.classified_by);
                if (actionContainer) {
                    actionContainer.style.display = isClassifier ? 'flex' : 'none';
                }
            }
        } else {

            if (bannerEl) bannerEl.style.display = 'none';
            if (promptContainer) {
                if (isNavahi) {
                    promptContainer.style.display = 'none';
                } else {
                    promptContainer.style.display = 'block';
                    const statusEl = document.getElementById('classificationPromptStatusMsg');
                    if (statusEl) statusEl.style.display = 'none';
                    const btnHelpful = document.getElementById('btnClassifyHelpful');
                    const btnUseless = document.getElementById('btnClassifyUseless');
                    if (btnHelpful) btnHelpful.disabled = false;
                    if (btnUseless) btnUseless.disabled = false;
                }
            }
        }
    }

    async function submitClassification(classification) {
        if (!currentViewingReportId) return;
        const btnHelpful = document.getElementById('btnClassifyHelpful');
        const btnUseless = document.getElementById('btnClassifyUseless');

        if (btnHelpful) btnHelpful.disabled = true;
        if (btnUseless) btnUseless.disabled = true;

        try {
            const res = await adminFetch('/api/admin/report/classify', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    report_id: currentViewingReportId,
                    classification: classification
                })
            });

            if (res.ok) {
                const respData = await res.json();
                showToast('ارزیابی کیفیت گزارش با موفقیت ثبت شد', 2500, 'success');

                const repInAll = allReports ? allReports.find(r => r.id === currentViewingReportId) : null;
                if (repInAll) repInAll.classification = classification;
                const repInFiltered = filteredReports ? filteredReports.find(r => r.id === currentViewingReportId) : null;
                if (repInFiltered) repInFiltered.classification = classification;

                const promptContainer = document.getElementById('classification-prompt-container');
                if (promptContainer) {
                    promptContainer.style.opacity = '0';
                    promptContainer.style.transform = 'translateY(10px)';
                    setTimeout(() => {
                        promptContainer.style.display = 'none';
                        promptContainer.style.opacity = '1';
                        promptContainer.style.transform = 'none';
                    }, 250);
                }

                updateClassificationUI({
                    classification: classification,
                    classified_by: respData.classified_by || (currentAdmin ? currentAdmin.username : ''),
                    classified_by_name: respData.classified_by_name || (currentAdmin ? currentAdmin.full_name : '')
                }, currentViewingReportId);

            } else {
                const errText = await res.text();
                showToast(errText || 'خطا در ثبت ارزیابی گزارش', 3500, 'error');
                if (btnHelpful) btnHelpful.disabled = false;
                if (btnUseless) btnUseless.disabled = false;
            }
        } catch (err) {
            showToast('خطای ارتباط با سرور', 3500, 'error');
            if (btnHelpful) btnHelpful.disabled = false;
            if (btnUseless) btnUseless.disabled = false;
        }
    }

    function setupClassificationModal() {
        const modal = document.getElementById('modal-change-classification');
        const btnOpen = document.getElementById('btnOpenChangeClassification');
        const btnClose = document.getElementById('btn-close-change-classification-modal');
        const btnCancel = document.getElementById('btn-cancel-change-classification');
        const form = document.getElementById('form-change-classification');
        const statusMsg = document.getElementById('changeClassificationStatusMsg');

        const radioHelpful = document.getElementById('radio-helpful');
        const radioUseless = document.getElementById('radio-useless');
        const cardHelpful = document.getElementById('option-card-helpful');
        const cardUseless = document.getElementById('option-card-useless');

        const updateCardSelected = () => {
            if (cardHelpful && radioHelpful) {
                cardHelpful.classList.toggle('selected', radioHelpful.checked);
            }
            if (cardUseless && radioUseless) {
                cardUseless.classList.toggle('selected', radioUseless.checked);
            }
        };

        radioHelpful?.addEventListener('change', updateCardSelected);
        radioUseless?.addEventListener('change', updateCardSelected);

        cardHelpful?.addEventListener('click', () => {
            if (radioHelpful) {
                radioHelpful.checked = true;
                updateCardSelected();
            }
        });

        cardUseless?.addEventListener('click', () => {
            if (radioUseless) {
                radioUseless.checked = true;
                updateCardSelected();
            }
        });

        btnOpen?.addEventListener('click', () => {
            if (!currentViewingReportId) return;
            const rep = allReports ? allReports.find(r => r.id === currentViewingReportId) : null;
            const currentVal = rep?.classification || 'helpful';

            if (radioHelpful && radioUseless) {
                radioHelpful.checked = currentVal === 'helpful';
                radioUseless.checked = currentVal === 'useless';
                updateCardSelected();
            }

            if (statusMsg) statusMsg.style.display = 'none';
            if (modal) modal.style.display = 'flex';
        });

        const closeModal = () => {
            if (modal) modal.style.display = 'none';
        };

        btnClose?.addEventListener('click', closeModal);
        btnCancel?.addEventListener('click', closeModal);
        modal?.addEventListener('click', (e) => {
            if (e.target === modal) closeModal();
        });

        form?.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (!currentViewingReportId) return;

            const selected = form.querySelector('input[name="modal_classification"]:checked')?.value;
            if (!selected) {
                showToast('لطفاً یک گزینه را انتخاب فرمایید', 3000, 'error');
                return;
            }

            const submitBtn = document.getElementById('btn-submit-change-classification');
            const spinner = submitBtn?.querySelector('.btn-spinner');
            const btnText = submitBtn?.querySelector('.btn-text');

            if (submitBtn) submitBtn.disabled = true;
            if (spinner) spinner.classList.remove('d-none');

            try {
                const res = await adminFetch('/api/admin/report/classify', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        report_id: currentViewingReportId,
                        classification: selected
                    })
                });

                if (res.ok) {
                    const respData = await res.json();
                    closeModal();
                    showToast('ارزیابی گزارش با موفقیت تغییر یافت', 2500, 'success');

                    const repInAll = allReports ? allReports.find(r => r.id === currentViewingReportId) : null;
                    if (repInAll) repInAll.classification = selected;
                    const repInFiltered = filteredReports ? filteredReports.find(r => r.id === currentViewingReportId) : null;
                    if (repInFiltered) repInFiltered.classification = selected;

                    updateClassificationUI({
                        classification: selected,
                        classified_by: respData.classified_by || (currentAdmin ? currentAdmin.username : ''),
                        classified_by_name: respData.classified_by_name || (currentAdmin ? currentAdmin.full_name : '')
                    }, currentViewingReportId);

                } else {
                    const errText = await res.text();
                    if (statusMsg) {
                        statusMsg.className = 'response-box badge-critical';
                        statusMsg.textContent = errText || 'خطا در تغییر ارزیابی پرونده';
                        statusMsg.style.display = 'block';
                    }
                }
            } catch (err) {
                if (statusMsg) {
                    statusMsg.className = 'response-box badge-critical';
                    statusMsg.textContent = 'خطای ارتباط با سرور';
                    statusMsg.style.display = 'block';
                }
            } finally {
                if (submitBtn) submitBtn.disabled = false;
                if (spinner) spinner.classList.add('d-none');
            }
        });

        document.getElementById('btnClassifyHelpful')?.addEventListener('click', () => submitClassification('helpful'));
        document.getElementById('btnClassifyUseless')?.addEventListener('click', () => submitClassification('useless'));
    }

    setupClassificationModal();

    function renderClassificationPieChart(classData) {
        const container = document.getElementById('chart-classification-pie');
        const tagEl = document.getElementById('chart-quality-total-tag');
        if (!container) return;

        const totalClassified = classData ? (classData.total_classified || 0) : 0;
        const helpfulCount = classData ? (classData.total_helpful || 0) : 0;
        const uselessCount = classData ? (classData.total_useless || 0) : 0;
        const helpfulPct = classData ? (classData.helpful_percentage || 0) : 0;
        const uselessPct = classData ? (classData.useless_percentage || 0) : 0;

        if (tagEl) {
            tagEl.textContent = `${toPersianNum(totalClassified)} پرونده ارزیابی‌شده`;
        }

        if (totalClassified === 0) {
            container.innerHTML = `
                <div class="chart-empty-state">
                    <div class="chart-empty-icon">
                        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="12" cy="12" r="10"></circle>
                            <path d="M12 2a10 10 0 0 1 10 10H12V2z"></path>
                        </svg>
                    </div>
                    <span>هنوز هیچ پرونده‌ای توسط کارشناسان ارزیابی کیفیت نشده است.</span>
                </div>
            `;
            return;
        }

        const size = 230;
        const strokeWidth = 24;
        const radius = (size / 2) - (strokeWidth / 2);
        const circumference = 2 * Math.PI * radius;

        const segments = [
            {
                key: 'helpful',
                value: helpfulCount,
                percentage: helpfulPct,
                color: '#10b981',
                label: 'گزارش‌های مفید و موثر',
                shortLabel: 'گزارش‌های مفید',
                pillClass: 'helpful'
            },
            {
                key: 'useless',
                value: uselessCount,
                percentage: uselessPct,
                color: '#64748b',
                label: 'گزارش‌های بی‌محتوا و فاقد ارزش',
                shortLabel: 'بی‌محتوا و فاقد ارزش',
                pillClass: 'useless'
            }
        ];

        const nonZeroSegments = segments.filter(s => s.value > 0);
        const bothPresent = nonZeroSegments.length > 1;
        const gap = bothPresent ? 14 : 0;

        let circlesHtml = '';
        let cumulativePct = 0;

        nonZeroSegments.forEach((seg) => {
            const rawLength = (seg.percentage / 100) * circumference;
            const arcLength = Math.max(2, rawLength - gap);
            const targetOffset = -((cumulativePct / 100) * circumference + (gap / 2));

            circlesHtml += `
                <circle
                    id="donut-seg-${seg.key}"
                    class="donut-segment"
                    cx="${size / 2}"
                    cy="${size / 2}"
                    r="${radius}"
                    fill="none"
                    stroke="${seg.color}"
                    stroke-dasharray="${arcLength} ${circumference}"
                    stroke-dashoffset="-${circumference}"
                    data-target-offset="${targetOffset}"
                    data-key="${seg.key}"
                    style="filter: none; pointer-events: stroke;"
                />
            `;
            cumulativePct += seg.percentage;
        });

        const legendHtml = segments.map((seg) => `
            <div class="donut-legend-item" id="donut-legend-${seg.key}" data-key="${seg.key}">
                <div class="donut-legend-item-right">
                    <span class="donut-legend-dot" style="background: ${seg.color}; box-shadow: 0 0 6px ${seg.color}80;"></span>
                    <span class="donut-legend-name">${seg.label}</span>
                </div>
                <div class="donut-legend-item-left">
                    <span class="donut-legend-pct-pill ${seg.pillClass}">${toPersianNum(seg.percentage)}٪</span>
                    <span class="donut-legend-val">${toPersianNum(seg.value)} پرونده</span>
                </div>
            </div>
        `).join('');

        container.innerHTML = `
            <div class="donut-card-container">
                <div class="donut-chart-wrapper" id="donutChartWrapper">
                    <svg class="donut-svg" viewBox="0 0 ${size} ${size}" aria-label="نمودار تفکیک کیفیت پرونده‌ها" role="img">
                        <!-- Background base track ring -->
                        <circle
                            class="donut-base-ring"
                            cx="${size / 2}"
                            cy="${size / 2}"
                            r="${radius}"
                            fill="none"
                            style="pointer-events: none;"
                        />
                        <!-- Data Segments -->
                        ${circlesHtml}
                    </svg>

                    <!-- Interactive Center Content -->
                    <div class="donut-center-content" id="donutCenterContent">
                        <span class="donut-center-label" id="donutCenterLabel">کل ارزیابی‌ها</span>
                        <span class="donut-center-value" id="donutCenterValue">${toPersianNum(totalClassified)}</span>
                        <span class="donut-center-pct" id="donutCenterPct" style="color: var(--text-muted);">پرونده</span>
                    </div>
                </div>

                <!-- Legend Breakdown List -->
                <div class="donut-legend-list">
                    ${legendHtml}
                </div>
            </div>
        `;

        requestAnimationFrame(() => {
            nonZeroSegments.forEach(seg => {
                const circleEl = document.getElementById(`donut-seg-${seg.key}`);
                if (circleEl) {
                    const targetOffset = circleEl.getAttribute('data-target-offset');
                    circleEl.style.strokeDashoffset = `${targetOffset}`;
                }
            });
        });

        const centerLabel = document.getElementById('donutCenterLabel');
        const centerValue = document.getElementById('donutCenterValue');
        const centerPct = document.getElementById('donutCenterPct');

        function setHoveredState(key) {
            const seg = segments.find(s => s.key === key);
            if (!seg) return;

            segments.forEach(s => {
                const circleEl = document.getElementById(`donut-seg-${s.key}`);
                const legendEl = document.getElementById(`donut-legend-${s.key}`);
                const isTarget = s.key === key;

                if (circleEl) {
                    if (isTarget) {
                        circleEl.classList.add('active');
                        circleEl.style.filter = `drop-shadow(0px 0px 8px ${s.color}) brightness(1.15)`;
                    } else {
                        circleEl.classList.remove('active');
                        circleEl.style.filter = 'none';
                    }
                }
                if (legendEl) {
                    legendEl.classList.toggle('active', isTarget);
                }
            });

            if (centerLabel) centerLabel.textContent = seg.shortLabel || seg.label;
            if (centerValue) centerValue.textContent = toPersianNum(seg.value);
            if (centerPct) {
                centerPct.textContent = `[${toPersianNum(seg.percentage)}٪]`;
                centerPct.style.color = seg.color;
            }
        }

        function resetHoveredState() {
            segments.forEach(s => {
                const circleEl = document.getElementById(`donut-seg-${s.key}`);
                const legendEl = document.getElementById(`donut-legend-${s.key}`);
                if (circleEl) {
                    circleEl.classList.remove('active');
                    circleEl.style.filter = 'none';
                }
                if (legendEl) {
                    legendEl.classList.remove('active');
                }
            });

            if (centerLabel) centerLabel.textContent = 'کل ارزیابی‌ها';
            if (centerValue) centerValue.textContent = toPersianNum(totalClassified);
            if (centerPct) {
                centerPct.textContent = 'پرونده';
                centerPct.style.color = 'var(--text-muted)';
            }
        }

        const donutWrapper = container.querySelector('.donut-chart-wrapper');
        if (donutWrapper) {
            donutWrapper.addEventListener('mouseleave', resetHoveredState);
        }

        nonZeroSegments.forEach(seg => {
            const circleEl = document.getElementById(`donut-seg-${seg.key}`);
            if (circleEl) {
                circleEl.addEventListener('mouseenter', () => setHoveredState(seg.key));
            }
        });

        const legendListEl = container.querySelector('.donut-legend-list');
        if (legendListEl) {
            legendListEl.addEventListener('mouseleave', resetHoveredState);
        }

        segments.forEach(seg => {
            const legendEl = document.getElementById(`donut-legend-${seg.key}`);
            if (legendEl) {
                legendEl.addEventListener('mouseenter', () => setHoveredState(seg.key));
            }
        });
    }

    function renderCategoryPieChart(categoryBreakdown, totalReports) {
        const container = document.getElementById('chart-categories');
        const tagEl = document.getElementById('chart-categories-total-tag');
        if (!container) return;

        const validItems = (categoryBreakdown || []).filter(item => item.count > 0);
        const total = totalReports || validItems.reduce((sum, item) => sum + item.count, 0) || 0;

        if (tagEl) {
            tagEl.textContent = `${toPersianNum(total)} پرونده`;
        }

        if (validItems.length === 0 || total === 0) {
            container.innerHTML = `
                <div class="chart-empty-state">
                    <div class="chart-empty-icon">
                        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M21.21 15.89A10 10 0 1 1 8 2.83"></path>
                            <path d="M22 12A10 10 0 0 0 12 2v10z"></path>
                        </svg>
                    </div>
                    <span>هنوز هیچ گزارشی برای نمایش توزیع موضوعی ثبت نشده است.</span>
                </div>
            `;
            return;
        }

        const palette = [
            '#38bdf8',
            '#34d399',
            '#818cf8',
            '#fbbf24',
            '#f43f5e',
            '#a78bfa',
            '#2dd4bf',
            '#fb923c',
            '#94a3b8'
        ];

        const size = 230;
        const strokeWidth = 24;
        const radius = (size / 2) - (strokeWidth / 2);
        const circumference = 2 * Math.PI * radius;

        const segments = validItems.map((item, idx) => {
            const pct = Math.round((item.count / total) * 1000) / 10;
            return {
                id: `cat-${idx}`,
                label: item.category || 'سایر',
                value: item.count,
                percentage: pct,
                color: palette[idx % palette.length]
            };
        });

        const bothPresent = segments.length > 1;
        const gap = bothPresent ? Math.min(12, Math.max(6, Math.floor(60 / segments.length))) : 0;

        let circlesHtml = '';
        let cumulativePct = 0;

        segments.forEach((seg) => {
            const rawLength = (seg.percentage / 100) * circumference;
            const arcLength = Math.max(2, rawLength - gap);
            const targetOffset = -((cumulativePct / 100) * circumference + (gap / 2));

            circlesHtml += `
                <circle
                    id="donut-cat-seg-${seg.id}"
                    class="donut-segment"
                    cx="${size / 2}"
                    cy="${size / 2}"
                    r="${radius}"
                    fill="none"
                    stroke="${seg.color}"
                    stroke-dasharray="${arcLength} ${circumference}"
                    stroke-dashoffset="-${circumference}"
                    data-target-offset="${targetOffset}"
                    data-id="${seg.id}"
                    style="filter: none; pointer-events: stroke;"
                />
            `;
            cumulativePct += seg.percentage;
        });

        const legendHtml = segments.map((seg) => `
            <div class="donut-legend-item" id="donut-cat-legend-${seg.id}" data-id="${seg.id}">
                <div class="donut-legend-item-right">
                    <span class="donut-legend-dot" style="background: ${seg.color}; box-shadow: 0 0 6px ${seg.color}80;"></span>
                    <span class="donut-legend-name">${escapeHTML(seg.label)}</span>
                </div>
                <div class="donut-legend-item-left">
                    <span class="donut-legend-pct-pill" style="background: ${seg.color}22; color: ${seg.color};">${toPersianNum(seg.percentage)}٪</span>
                    <span class="donut-legend-val">${toPersianNum(seg.value)} پرونده</span>
                </div>
            </div>
        `).join('');

        container.innerHTML = `
            <div class="donut-card-container">
                <div class="donut-chart-wrapper">
                    <svg class="donut-svg" viewBox="0 0 ${size} ${size}" aria-label="نمودار توزیع موضوعی گزارش‌ها" role="img">
                        <!-- Background track ring -->
                        <circle
                            class="donut-base-ring"
                            cx="${size / 2}"
                            cy="${size / 2}"
                            r="${radius}"
                            fill="none"
                            style="pointer-events: none;"
                        />
                        <!-- Data Segments -->
                        ${circlesHtml}
                    </svg>

                    <!-- Interactive Center Content -->
                    <div class="donut-center-content" id="donutCatCenterContent">
                        <span class="donut-center-label" id="donutCatCenterLabel">کل گزارش‌ها</span>
                        <span class="donut-center-value" id="donutCatCenterValue">${toPersianNum(total)}</span>
                        <span class="donut-center-pct" id="donutCatCenterPct" style="color: var(--text-muted);">پرونده</span>
                    </div>
                </div>

                <!-- Legend Breakdown List -->
                <div class="donut-legend-list">
                    ${legendHtml}
                </div>
            </div>
        `;

        requestAnimationFrame(() => {
            segments.forEach(seg => {
                const circleEl = document.getElementById(`donut-cat-seg-${seg.id}`);
                if (circleEl) {
                    const targetOffset = circleEl.getAttribute('data-target-offset');
                    circleEl.style.strokeDashoffset = `${targetOffset}`;
                }
            });
        });

        const centerLabel = document.getElementById('donutCatCenterLabel');
        const centerValue = document.getElementById('donutCatCenterValue');
        const centerPct = document.getElementById('donutCatCenterPct');

        function setHoveredCat(segId) {
            const seg = segments.find(s => s.id === segId);
            if (!seg) return;

            segments.forEach(s => {
                const circleEl = document.getElementById(`donut-cat-seg-${s.id}`);
                const legendEl = document.getElementById(`donut-cat-legend-${s.id}`);
                const isTarget = s.id === segId;

                if (circleEl) {
                    if (isTarget) {
                        circleEl.classList.add('active');
                        circleEl.style.filter = `drop-shadow(0px 0px 8px ${s.color}) brightness(1.15)`;
                    } else {
                        circleEl.classList.remove('active');
                        circleEl.style.filter = 'none';
                    }
                }
                if (legendEl) {
                    legendEl.classList.toggle('active', isTarget);
                }
            });

            if (centerLabel) centerLabel.textContent = seg.label;
            if (centerValue) centerValue.textContent = toPersianNum(seg.value);
            if (centerPct) {
                centerPct.textContent = `[${toPersianNum(seg.percentage)}٪]`;
                centerPct.style.color = seg.color;
            }
        }

        function resetHoveredCat() {
            segments.forEach(s => {
                const circleEl = document.getElementById(`donut-cat-seg-${s.id}`);
                const legendEl = document.getElementById(`donut-cat-legend-${s.id}`);
                if (circleEl) {
                    circleEl.classList.remove('active');
                    circleEl.style.filter = 'none';
                }
                if (legendEl) {
                    legendEl.classList.remove('active');
                }
            });

            if (centerLabel) centerLabel.textContent = 'کل گزارش‌ها';
            if (centerValue) centerValue.textContent = toPersianNum(total);
            if (centerPct) {
                centerPct.textContent = 'پرونده';
                centerPct.style.color = 'var(--text-muted)';
            }
        }

        const donutWrapper = container.querySelector('.donut-chart-wrapper');
        if (donutWrapper) {
            donutWrapper.addEventListener('mouseleave', resetHoveredCat);
        }

        segments.forEach(seg => {
            const circleEl = document.getElementById(`donut-cat-seg-${seg.id}`);
            if (circleEl) {
                circleEl.addEventListener('mouseenter', () => setHoveredCat(seg.id));
            }

            const legendEl = document.getElementById(`donut-cat-legend-${seg.id}`);
            if (legendEl) {
                legendEl.addEventListener('mouseenter', () => setHoveredCat(seg.id));
            }
        });

        const legendListEl = container.querySelector('.donut-legend-list');
        if (legendListEl) {
            legendListEl.addEventListener('mouseleave', resetHoveredCat);
        }
    }

    function renderPrioritiesSegmentedBar(priorityBreakdown, totalReports) {
        const container = document.getElementById('chart-priorities');
        const tagEl = document.getElementById('chart-priorities-total-tag');
        if (!container) return;

        const critItem = (priorityBreakdown || []).find(p => p.priority === 'critical');
        const urgItem = (priorityBreakdown || []).find(p => p.priority === 'urgent');
        const normItem = (priorityBreakdown || []).find(p => p.priority === 'normal');

        const critCount = critItem ? critItem.count : 0;
        const urgCount = urgItem ? urgItem.count : 0;
        const normCount = normItem ? normItem.count : 0;
        const total = totalReports || (critCount + urgCount + normCount) || 0;

        if (tagEl) {
            tagEl.textContent = `${toPersianNum(total)} پرونده`;
        }

        if (total === 0) {
            container.innerHTML = `
                <div class="chart-empty-state">
                    <div class="chart-empty-icon">
                        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
                            <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
                        </svg>
                    </div>
                    <span>هنوز هیچ اطلاعاتی برای تفکیک فوریت پرونده‌ها ثبت نشده است.</span>
                </div>
            `;
            return;
        }

        const critPct = Math.round((critCount / total) * 1000) / 10;
        const urgPct = Math.round((urgCount / total) * 1000) / 10;
        const normPct = Math.max(0, Math.round((100 - critPct - urgPct) * 10) / 10);

        let overallBadge = '';
        if (critCount > 0) {
            overallBadge = `
                <span class="priority-overall-badge critical">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                        <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
                    </svg>
                    <span>${toPersianNum(critCount)} پرونده ویژه و آنی</span>
                </span>
            `;
        } else if (urgCount > 0) {
            overallBadge = `
                <span class="priority-overall-badge urgent">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                        <circle cx="12" cy="12" r="10"></circle>
                        <polyline points="12 6 12 12 16 14"></polyline>
                    </svg>
                    <span>${toPersianNum(urgCount)} پرونده فوری</span>
                </span>
            `;
        } else {
            overallBadge = `
                <span class="priority-overall-badge normal">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                        <polyline points="22 4 12 14.01 9 11.01"></polyline>
                    </svg>
                    <span>وضعیت عادی و پایدار</span>
                </span>
            `;
        }

        const prioItems = [
            {
                key: 'critical',
                label: 'آنی و ویژه (حداکثر حساسیت)',
                shortLabel: 'آنی و ویژه',
                count: critCount,
                pct: critPct,
                color: '#f43f5e',
                iconSvg: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>`
            },
            {
                key: 'urgent',
                label: 'فوری (رسیدگی با اولویت بالا)',
                shortLabel: 'فوری',
                count: urgCount,
                pct: urgPct,
                color: '#f59e0b',
                iconSvg: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>`
            },
            {
                key: 'normal',
                label: 'عادی (روال معمول اداری)',
                shortLabel: 'عادی',
                count: normCount,
                pct: normPct,
                color: '#10b981',
                iconSvg: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>`
            }
        ];

        let barSegmentsHtml = '';
        prioItems.forEach(item => {
            if (item.count > 0) {
                barSegmentsHtml += `
                    <div
                        class="segmented-progress-bar-item"
                        id="prio-bar-${item.key}"
                        data-key="${item.key}"
                        style="width: 0%; background: ${item.color};"
                        data-target-width="${item.pct}%"
                        title="${item.shortLabel}: ${toPersianNum(item.pct)}٪ (${toPersianNum(item.count)} پرونده)"
                    ></div>
                `;
            }
        });

        let legendRowsHtml = prioItems.map(item => `
            <div class="donut-legend-item" id="prio-legend-${item.key}" data-key="${item.key}">
                <div class="donut-legend-item-right">
                    <span class="prio-icon-badge ${item.key}">${item.iconSvg}</span>
                    <span class="donut-legend-name">${item.label}</span>
                </div>
                <div class="donut-legend-item-left">
                    <span class="donut-legend-pct-pill" style="background: ${item.color}22; color: ${item.color};">${toPersianNum(item.pct)}٪</span>
                    <span class="donut-legend-val">${toPersianNum(item.count)} پرونده</span>
                </div>
            </div>
        `).join('');

        container.innerHTML = `
            <div class="priority-segmented-card">
                <div class="priority-overall-header">
                    <span class="priority-overall-title">پایش فوریت و ضرورت رسیدگی</span>
                    ${overallBadge}
                </div>

                <div class="segmented-progress-track" role="progressbar" aria-label="توزیع اولویت و فوریت پرونده‌ها" aria-valuemin="0" aria-valuemax="100">
                    ${barSegmentsHtml}
                </div>

                <div class="donut-legend-list" style="margin-top: 10px;">
                    ${legendRowsHtml}
                </div>
            </div>
        `;

        requestAnimationFrame(() => {
            prioItems.forEach(item => {
                const el = document.getElementById(`prio-bar-${item.key}`);
                if (el) {
                    el.style.width = el.getAttribute('data-target-width');
                }
            });
        });

        prioItems.forEach(item => {
            const barEl = document.getElementById(`prio-bar-${item.key}`);
            const legendEl = document.getElementById(`prio-legend-${item.key}`);

            function highlight() {
                if (barEl) {
                    barEl.classList.add('active');
                    barEl.style.filter = `brightness(1.25) drop-shadow(0 0 6px ${item.color})`;
                }
                if (legendEl) legendEl.classList.add('active');
            }

            function unhighlight() {
                if (barEl) {
                    barEl.classList.remove('active');
                    barEl.style.filter = '';
                }
                if (legendEl) legendEl.classList.remove('active');
            }

            if (barEl) {
                barEl.addEventListener('mouseenter', highlight);
                barEl.addEventListener('mouseleave', unhighlight);
            }
            if (legendEl) {
                legendEl.addEventListener('mouseenter', highlight);
                legendEl.addEventListener('mouseleave', unhighlight);
            }
        });
    }

    let cachedAnalyticsData = null;
    let lastAnalyticsFetchTime = 0;
    const ANALYTICS_CACHE_TTL = 60000;

    function populateAnalyticsDOM(data) {

        const totalEl = document.getElementById('ana-total-reports');
        const respEl = document.getElementById('ana-total-responded');
        const unreadEl = document.getElementById('ana-total-unread');
        const avgEl = document.getElementById('ana-avg-response');

        const helpfulEl = document.getElementById('ana-total-helpful');
        const uselessEl = document.getElementById('ana-total-useless');

        if (totalEl) totalEl.textContent = toPersianNum(data.total_reports || 0);
        if (respEl) respEl.textContent = toPersianNum(data.total_responded || 0);
        if (unreadEl) unreadEl.textContent = toPersianNum(data.total_unread || 0);
        if (avgEl) {
            const hours = Math.round((data.avg_response_hours || 0) * 10) / 10;
            avgEl.textContent = hours > 24 ? `${toPersianNum(Math.round(hours / 24))} روز` : `${toPersianNum(hours)} ساعت`;
        }

        if (data.classification_analytics) {
            const ca = data.classification_analytics;
            if (helpfulEl) helpfulEl.innerHTML = `${toPersianNum(ca.total_helpful)} <span class="metric-sub-pct" id="ana-helpful-pct">(${toPersianNum(ca.helpful_percentage)}٪)</span>`;
            if (uselessEl) uselessEl.innerHTML = `${toPersianNum(ca.total_useless)} <span class="metric-sub-pct" id="ana-useless-pct">(${toPersianNum(ca.useless_percentage)}٪)</span>`;
            renderClassificationPieChart(ca);
        } else {
            renderClassificationPieChart(null);
        }

        renderCategoryPieChart(data.category_breakdown, data.total_reports);

        renderPrioritiesSegmentedBar(data.priority_breakdown, data.total_reports);

        renderMonthlyChart(data.monthly_trend);
    }

    async function renderAnalyticsDashboard(force = false) {
        const anaSkeleton = document.getElementById('analytics-skeleton-content');
        const anaActual = document.getElementById('analytics-actual-content');

        if (!force && cachedAnalyticsData && (Date.now() - lastAnalyticsFetchTime < ANALYTICS_CACHE_TTL)) {
            if (anaSkeleton) anaSkeleton.classList.add('d-none');
            if (anaActual) anaActual.classList.remove('d-none');
            populateAnalyticsDOM(cachedAnalyticsData);
            return true;
        }

        if (anaSkeleton && anaActual && !cachedAnalyticsData) {
            anaSkeleton.classList.remove('d-none');
            anaActual.classList.add('d-none');
        }

        try {
            const res = await adminFetch('/api/admin/analytics');
            if (!res.ok) {
                if (anaSkeleton) anaSkeleton.classList.add('d-none');
                if (anaActual) anaActual.classList.remove('d-none');
                renderClientSideAnalytics();
                throw new Error(`خطای دریافت آمار: ${res.status}`);
            }
            const data = await res.json();
            cachedAnalyticsData = data;
            lastAnalyticsFetchTime = Date.now();

            populateAnalyticsDOM(data);

            if (anaSkeleton) anaSkeleton.classList.add('d-none');
            if (anaActual) anaActual.classList.remove('d-none');
            return true;

        } catch (e) {
            if (anaSkeleton) anaSkeleton.classList.add('d-none');
            if (anaActual) anaActual.classList.remove('d-none');
            renderClientSideAnalytics();
            throw e;
        }
    }

    function renderMonthlyChart(monthlyTrend) {
        const monthNames = ['فروردین', 'اردیبهشت', 'خرداد', 'تیر', 'مرداد', 'شهریور', 'مهر', 'آبان', 'آذر', 'دی', 'بهمن', 'اسفند'];
        const monthCounts = new Array(12).fill(0);

        if (Array.isArray(monthlyTrend) && monthlyTrend.length === 12) {
            for (let i = 0; i < 12; i++) {
                monthCounts[i] = Number(monthlyTrend[i]) || 0;
            }
        } else if (allReports && allReports.length > 0) {
            for (let i = 0; i < allReports.length; i++) {
                const r = allReports[i];
                if (r._shamsiMonth) {
                    const mIdx = parseInt(r._shamsiMonth, 10) - 1;
                    if (mIdx >= 0 && mIdx < 12) {
                        monthCounts[mIdx]++;
                    }
                }
            }
        }

        const monthlyContainer = document.getElementById('chart-monthly');
        if (!monthlyContainer) return;

        const totalReports = monthCounts.reduce((acc, val) => acc + val, 0);
        const avgMonthly = totalReports > 0 ? (totalReports / 12).toFixed(1) : '۰';
        const maxVal = Math.max(...monthCounts, 0);
        const minVal = Math.min(...monthCounts);
        const peakIdx = maxVal > 0 ? monthCounts.indexOf(maxVal) : -1;
        const lowestIdx = monthCounts.indexOf(minVal);
        const peakMonthText = peakIdx >= 0 ? `${monthNames[peakIdx]} (${toPersianNum(maxVal)})` : 'ثبت‌نشده';
        const lowestMonthText = `${monthNames[lowestIdx]} (${toPersianNum(minVal)})`;
        const activeMonths = monthCounts.filter(c => c > 0).length;

        const svgWidth = 960;
        const svgHeight = 240;
        const padX = 55;
        const yTop = 26;
        const yBottom = 195;
        const effectiveMax = maxVal > 0 ? Math.ceil(maxVal * 1.25) : 4;
        const stepX = (svgWidth - padX * 2) / 11;

        const points = monthNames.map((name, i) => {
            const count = monthCounts[i];
            const ratio = count / effectiveMax;
            const x = (svgWidth - padX) - (i * stepX);
            const y = yBottom - (ratio * (yBottom - yTop));
            return { x, y, count, name, index: i };
        });

        function generateSpline(pts) {
            if (!pts || pts.length === 0) return '';
            if (pts.length === 1) return `M ${pts[0].x.toFixed(1)} ${pts[0].y.toFixed(1)}`;
            let d = `M ${pts[0].x.toFixed(1)} ${pts[0].y.toFixed(1)}`;
            for (let i = 0; i < pts.length - 1; i++) {
                const p0 = pts[i === 0 ? i : i - 1];
                const p1 = pts[i];
                const p2 = pts[i + 1];
                const p3 = pts[i + 2 < pts.length ? i + 2 : i + 1];

                const cp1x = p1.x + (p2.x - p0.x) / 6;
                const cp1y = p1.y + (p2.y - p0.y) / 6;
                const cp2x = p2.x - (p3.x - p1.x) / 6;
                const cp2y = p2.y - (p3.y - p1.y) / 6;

                d += ` C ${cp1x.toFixed(1)} ${cp1y.toFixed(1)}, ${cp2x.toFixed(1)} ${cp2y.toFixed(1)}, ${p2.x.toFixed(1)} ${p2.y.toFixed(1)}`;
            }
            return d;
        }

        const linePath = generateSpline(points);
        const areaPath = `${linePath} L ${points[11].x.toFixed(1)} ${yBottom} L ${points[0].x.toFixed(1)} ${yBottom} Z`;

        const gridLevels = [0, 0.33, 0.66, 1];
        const gridLinesHtml = gridLevels.map(lvl => {
            const y = yBottom - lvl * (yBottom - yTop);
            const val = Math.round(lvl * effectiveMax);
            return `
                <line x1="40" y1="${y.toFixed(1)}" x2="${svgWidth - 40}" y2="${y.toFixed(1)}" stroke="rgba(255, 255, 255, 0.07)" stroke-dasharray="4 6" />
                <text x="${svgWidth - 32}" y="${(y + 4).toFixed(1)}" font-size="11" fill="#64748b" text-anchor="start" font-family="var(--font-main, 'Vazirmatn', sans-serif)">${toPersianNum(val)}</text>
            `;
        }).join('');

        const dotsHtml = points.map(p => {
            const isPeak = p.count === maxVal && maxVal > 0;
            if (isPeak) {
                return `
                    <circle cx="${p.x.toFixed(1)}" cy="${p.y.toFixed(1)}" r="8" fill="none" stroke="#38bdf8" stroke-width="1.5" class="pulsing-halo" />
                    <circle cx="${p.x.toFixed(1)}" cy="${p.y.toFixed(1)}" r="5" fill="#38bdf8" stroke="#ffffff" stroke-width="2" filter="url(#monthlyDotShadow)" />
                `;
            }
            if (p.count > 0) {
                return `
                    <circle cx="${p.x.toFixed(1)}" cy="${p.y.toFixed(1)}" r="3.5" fill="#38bdf8" stroke="#0f172a" stroke-width="1.5" />
                `;
            }
            return `
                <circle cx="${p.x.toFixed(1)}" cy="${p.y.toFixed(1)}" r="2" fill="rgba(255, 255, 255, 0.18)" />
            `;
        }).join('');

        const xLabelsHtml = points.map((p, i) => {
            const isPeak = p.count === maxVal && maxVal > 0;
            const rightPct = ((svgWidth - p.x) / svgWidth * 100).toFixed(2);
            return `
                <div class="monthly-xaxis-item ${isPeak ? 'peak' : ''}" style="right: ${rightPct}%;">
                    <span class="monthly-xaxis-text">${p.name}</span>
                </div>
            `;
        }).join('');

        monthlyContainer.innerHTML = `
            <div class="monthly-chart-root">
                <!-- Top Summary & Stats Row -->
                <div class="monthly-stats-header">
                    <div class="monthly-primary-stat">
                        <span class="monthly-stat-title">کل گزارش‌های ثبت‌شده سال جاری</span>
                        <div class="monthly-stat-val-row">
                            <span class="monthly-stat-number">${toPersianNum(totalReports)}</span>
                            <span class="monthly-stat-unit">پرونده</span>
                            <div class="monthly-trend-pill">
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                                    <polyline points="23 6 13.5 15.5 8.5 10.5 1 18"></polyline>
                                    <polyline points="17 6 23 6 23 12"></polyline>
                                </svg>
                                <span>میانگین ماهانه: ${toPersianNum(avgMonthly)}</span>
                            </div>
                        </div>
                    </div>

                    <div class="monthly-secondary-stats">
                        <div class="monthly-substat">
                            <span class="substat-label">بیشترین ثبت:</span>
                            <span class="substat-value high">${peakMonthText}</span>
                        </div>
                        <div class="monthly-substat-divider"></div>
                        <div class="monthly-substat">
                            <span class="substat-label">کمترین ثبت:</span>
                            <span class="substat-value low">${lowestMonthText}</span>
                        </div>
                        <div class="monthly-substat-divider"></div>
                        <div class="monthly-substat">
                            <span class="substat-label">ماه‌های فعال:</span>
                            <span class="substat-value">${toPersianNum(activeMonths)} از ۱۲ ماه</span>
                        </div>
                    </div>
                </div>

                <!-- Chart Canvas Area with Dot Grid & Spline Area -->
                <div class="monthly-chart-canvas-wrap" id="monthlyChartCanvasWrap">
                    <svg class="monthly-spline-svg" viewBox="0 0 ${svgWidth} ${svgHeight}" preserveAspectRatio="none">
                        <defs>
                            <pattern id="monthlyDotGrid" x="0" y="0" width="22" height="22" patternUnits="userSpaceOnUse">
                                <circle cx="11" cy="11" r="1.1" fill="rgba(255, 255, 255, 0.08)" />
                            </pattern>
                            <linearGradient id="monthlyAreaGradient" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="0%" stop-color="#38bdf8" stop-opacity="0.30" />
                                <stop offset="70%" stop-color="#38bdf8" stop-opacity="0.05" />
                                <stop offset="100%" stop-color="#38bdf8" stop-opacity="0" />
                            </linearGradient>
                            <filter id="monthlyLineShadow" x="-10%" y="-20%" width="120%" height="150%">
                                <feDropShadow dx="0" dy="4" stdDeviation="6" flood-color="#0284c7" flood-opacity="0.65" />
                            </filter>
                            <filter id="monthlyDotShadow" x="-50%" y="-50%" width="200%" height="200%">
                                <feDropShadow dx="0" dy="2" stdDeviation="3" flood-color="rgba(0, 0, 0, 0.85)" />
                            </filter>
                        </defs>

                        <!-- Dot Grid Background -->
                        <rect x="0" y="0" width="100%" height="100%" fill="url(#monthlyDotGrid)" style="pointer-events: none;" />

                        <!-- Cartesian Horizontal Grid Lines & Ticks -->
                        ${gridLinesHtml}

                        <!-- Active Reference Tracking Line -->
                        <line id="monthlyRefLine" x1="0" y1="20" x2="0" y2="${yBottom}" stroke="#38bdf8" stroke-dasharray="4 4" stroke-width="1.5" style="display: none; pointer-events: none; opacity: 0.75;" />

                        <!-- Smooth Spline Gradient Area -->
                        <path class="monthly-spline-area" d="${areaPath}" fill="url(#monthlyAreaGradient)" />

                        <!-- Smooth Spline Line -->
                        <path class="monthly-spline-line" d="${linePath}" fill="none" stroke="#38bdf8" stroke-width="2.5" filter="url(#monthlyLineShadow)" />

                        <!-- Data Dots -->
                        ${dotsHtml}

                        <!-- Active Hover Dot Marker -->
                        <circle id="monthlyActiveDot" r="6.5" fill="#38bdf8" stroke="#ffffff" stroke-width="2.5" filter="url(#monthlyDotShadow)" style="display: none; pointer-events: none;" />
                    </svg>

                    <!-- X-Axis Labels (Persian Months RTL) -->
                    <div class="monthly-xaxis-bar">
                        ${xLabelsHtml}
                    </div>

                    <!-- Floating Interactive Tooltip -->
                    <div id="monthlyTooltip" class="monthly-custom-tooltip" style="display: none;">
                        <div class="monthly-tooltip-date" id="monthlyTooltipDate">فروردین</div>
                        <div class="monthly-tooltip-val-row">
                            <span class="monthly-tooltip-val" id="monthlyTooltipVal">۰ پرونده</span>
                            <span class="monthly-tooltip-pct" id="monthlyTooltipPct">[۰٪]</span>
                        </div>
                        <div class="monthly-tooltip-badge" id="monthlyTooltipBadge">ثبت‌شده در این ماه</div>
                    </div>
                </div>
            </div>
        `;

        const canvasWrap = document.getElementById('monthlyChartCanvasWrap');
        const refLine = document.getElementById('monthlyRefLine');
        const activeDot = document.getElementById('monthlyActiveDot');
        const tooltip = document.getElementById('monthlyTooltip');
        const tooltipDate = document.getElementById('monthlyTooltipDate');
        const tooltipVal = document.getElementById('monthlyTooltipVal');
        const tooltipPct = document.getElementById('monthlyTooltipPct');
        const tooltipBadge = document.getElementById('monthlyTooltipBadge');

        if (canvasWrap) {
            let chartMoveRaf = null;
            let lastEvent = null;

            canvasWrap.addEventListener('mousemove', (e) => {
                lastEvent = e;
                if (chartMoveRaf) return;
                chartMoveRaf = requestAnimationFrame(() => {
                    chartMoveRaf = null;
                    if (!lastEvent) return;
                    const ev = lastEvent;
                    const rect = canvasWrap.getBoundingClientRect();
                    const mouseX = ev.clientX - rect.left;
                    const svgX = (mouseX / rect.width) * svgWidth;

                    let nearestPt = points[0];
                    let minDiff = Infinity;
                    for (let i = 0; i < points.length; i++) {
                        const pt = points[i];
                        const diff = Math.abs(pt.x - svgX);
                        if (diff < minDiff) {
                            minDiff = diff;
                            nearestPt = pt;
                        }
                    }

                    if (refLine) {
                        refLine.style.display = 'block';
                        refLine.setAttribute('x1', nearestPt.x.toFixed(1));
                        refLine.setAttribute('x2', nearestPt.x.toFixed(1));
                    }

                    if (activeDot) {
                        activeDot.style.display = 'block';
                        activeDot.setAttribute('cx', nearestPt.x.toFixed(1));
                        activeDot.setAttribute('cy', nearestPt.y.toFixed(1));
                    }

                    if (tooltip) {
                        tooltip.style.display = 'block';
                        if (tooltipDate) tooltipDate.textContent = `${nearestPt.name} ماه`;
                        if (tooltipVal) tooltipVal.textContent = `${toPersianNum(nearestPt.count)} پرونده`;
                        const pct = totalReports > 0 ? ((nearestPt.count / totalReports) * 100).toFixed(1) : '۰';
                        if (tooltipPct) tooltipPct.textContent = `[${toPersianNum(pct)}٪]`;
                        if (tooltipBadge) {
                            if (nearestPt.count === maxVal && maxVal > 0) {
                                tooltipBadge.textContent = 'بیشترین ثبت سال جاری';
                                tooltipBadge.style.color = '#38bdf8';
                                tooltipBadge.style.background = 'rgba(56, 189, 248, 0.16)';
                            } else if (nearestPt.count > (totalReports / 12)) {
                                tooltipBadge.textContent = 'بالاتر از میانگین ماهانه';
                                tooltipBadge.style.color = '#34d399';
                                tooltipBadge.style.background = 'rgba(16, 185, 129, 0.16)';
                            } else if (nearestPt.count === 0) {
                                tooltipBadge.textContent = 'بدون ثبت گزارش';
                                tooltipBadge.style.color = '#94a3b8';
                                tooltipBadge.style.background = 'rgba(255, 255, 255, 0.06)';
                            } else {
                                tooltipBadge.textContent = 'ثبت‌شده در این ماه';
                                tooltipBadge.style.color = '#93c5fd';
                                tooltipBadge.style.background = 'rgba(59, 130, 246, 0.12)';
                            }
                        }

                        const ptClientX = (nearestPt.x / svgWidth) * rect.width;
                        const ptClientY = (nearestPt.y / svgHeight) * rect.height;
                        const clampedX = Math.max(75, Math.min(rect.width - 75, ptClientX));
                        tooltip.style.left = `${clampedX}px`;
                        tooltip.style.top = `${Math.max(12, ptClientY)}px`;
                    }
                });
            }, { passive: true });

            canvasWrap.addEventListener('mouseleave', () => {
                if (chartMoveRaf) {
                    cancelAnimationFrame(chartMoveRaf);
                    chartMoveRaf = null;
                }
                lastEvent = null;
                if (refLine) refLine.style.display = 'none';
                if (activeDot) activeDot.style.display = 'none';
                if (tooltip) tooltip.style.display = 'none';
            }, { passive: true });
        }
    }

    function renderClientSideAnalytics() {
        if (!allReports || allReports.length === 0) return;
        renderMonthlyChart();
    }

    document.getElementById('btnRefreshAnalytics')?.addEventListener('click', async () => {
        try {
            await renderAnalyticsDashboard(true);
            showToast('نمودارهای آماری بروزرسانی شدند', 2500, 'success');
        } catch (err) {
            showToast('خطا در دریافت اطلاعات آماری. ارتباط با سرور برقرار نشد.', 3500, 'error');
        }
    });

    function connectAdminSSE() {
        if (_sseConnection || typeof EventSource === 'undefined') return;
        if (!dashboardSection || dashboardSection.style.display === 'none' || dashboardSection.classList.contains('d-none') || !currentAdmin) return;
        _sseConnection = new EventSource('/api/admin/events');

        _sseConnection.addEventListener('new_report', () => {
            scheduleReportsRefresh();
        });

        _sseConnection.addEventListener('new_message', (e) => {
            try {
                const data = JSON.parse(e.data);
                markReportHasNewMessage(data.report_id);
            } catch (_) { }
            scheduleReportsRefresh();
        });

        _sseConnection.addEventListener('regional_response', () => {
            scheduleReportsRefresh();
        });

        _sseConnection.addEventListener('session_revoked', () => {
            triggerConcurrentLoginAlert();
        });

        _sseConnection.addEventListener('heartbeat', () => {

        });

        _sseConnection.addEventListener('error', () => {
            if (_sseConnection && _sseConnection.readyState === EventSource.CLOSED) {
                _sseConnection = null;
                startFallbackPolling();
            }
        });
    }

    function scheduleReportsRefresh() {
        clearTimeout(_refreshDebounceTimer);
        _refreshDebounceTimer = setTimeout(() => {
            fetchReports(true, false).catch(() => { });
            fetchNotifications().catch(() => { });
        }, 200);
    }

    function markReportHasNewMessage(reportId) {
        if (!reportId) return;
        const rep = allReports.find(r => r.id === parseInt(reportId, 10));
        if (rep) rep.has_new_message = true;
        const tr = reportsBody ? reportsBody.querySelector(`tr[data-report-id="${reportId}"]`) : null;
        if (tr) tr.classList.add('row-new-message');
    }

    function startFallbackPolling() {
        if (!dashboardSection || dashboardSection.style.display === 'none' || dashboardSection.classList.contains('d-none') || !currentAdmin) return;
        if (reportPollingInterval) return;
        reportPollingInterval = setInterval(() => {
            if (document.hidden) return;
            if (!dashboardSection || dashboardSection.style.display === 'none' || dashboardSection.classList.contains('d-none') || !currentAdmin) return;
            fetchReports(false, false).catch(() => { });
        }, 15000);
        notifPollingInterval = setInterval(() => {
            if (document.hidden) return;
            if (!dashboardSection || dashboardSection.style.display === 'none' || dashboardSection.classList.contains('d-none') || !currentAdmin) return;
            fetchNotifications().catch(() => { });
        }, 15000);
    }

    function startBackgroundPolling() {
        if (!dashboardSection || dashboardSection.style.display === 'none' || dashboardSection.classList.contains('d-none') || !currentAdmin) return;
        connectAdminSSE();
        if (!reportPollingInterval) {
            reportPollingInterval = setInterval(() => {
                if (document.hidden) return;
                if (!dashboardSection || dashboardSection.style.display === 'none' || dashboardSection.classList.contains('d-none') || !currentAdmin) return;
                fetchReports(false, false).catch(() => { });
            }, 15000);
        }
        if (!notifPollingInterval) {
            notifPollingInterval = setInterval(() => {
                if (document.hidden) return;
                if (!dashboardSection || dashboardSection.style.display === 'none' || dashboardSection.classList.contains('d-none') || !currentAdmin) return;
                fetchNotifications().catch(() => { });
            }, 15000);
        }
    }

    function stopBackgroundPolling() {
        clearInterval(reportPollingInterval);
        clearInterval(notifPollingInterval);
        reportPollingInterval = null;
        notifPollingInterval = null;
        if (_sseConnection) {
            _sseConnection.close();
            _sseConnection = null;
        }
    }

    document.addEventListener('visibilitychange', () => {
        if (!document.hidden && dashboardSection && dashboardSection.style.display !== 'none' && !dashboardSection.classList.contains('d-none') && currentAdmin) {
            fetchReports(false, false).catch(() => { });
            fetchNotifications().catch(() => { });
        }
    });

    async function checkAdminProfile() {
        try {
            const res = await adminFetch('/api/admin/me');
            if (res.ok) {
                currentAdmin = await res.json();
                if (currentAdmin && currentAdmin.csrf_token) {
                    sessionStorage.setItem('herasat_csrf', currentAdmin.csrf_token);
                }
                if (currentAdmin && (currentAdmin.server_time_unix_ms || currentAdmin.server_time) && typeof window.__syncServerTime === 'function') {
                    window.__syncServerTime(currentAdmin.server_time_unix_ms || currentAdmin.server_time);
                }
                setupSuperuserUI();
                checkMustChangePasswordPolicy();
                check2FAEnforcementNotice();
                return true;
            }
        } catch (e) {
            console.error("Failed to check profile", e);
        }
        return false;
    }

    function checkMustChangePasswordPolicy() {
        if (!currentAdmin) return;
        const modalForcePwd = document.getElementById('modal-force-change-password');
        const reasonText = document.getElementById('force-pwd-reason-text');

        if (currentAdmin.must_change_password) {
            if (modalForcePwd) modalForcePwd.style.display = 'flex';
            if (reasonText) {
                if (currentAdmin.is_password_expired) {
                    reasonText.textContent = 'هشدار امنیتی: مدت زمان اعتبار ۳۰ روزه رمز عبور شما به پایان رسیده است. طبق خط‌مشی امنیتی سامانه، تعیین رمز عبور جدید پیش از ورود به کارتابل الزامی می‌باشد.';
                } else {
                    reasonText.textContent = 'توجه: تغییر رمز عبور حساب کاربری شما توسط پشتیبانی سامانه الزامی شده است. لطفاً کلمه عبور جدید و امن خود را تعیین فرمایید.';
                }
            }
        } else {
            if (modalForcePwd) modalForcePwd.style.display = 'none';
        }
    }

    function check2FAEnforcementNotice() {
        if (!currentAdmin || currentAdmin.totp_enabled) return;
        if (currentAdmin.must_change_password) return;
        if (sessionStorage.getItem('2fa_prompt_shown') === 'true') return;
        sessionStorage.setItem('2fa_prompt_shown', 'true');

        setTimeout(async () => {
            if (!dashboardSection || dashboardSection.style.display === 'none' || dashboardSection.classList.contains('d-none')) {
                return;
            }
            if (!currentAdmin || currentAdmin.totp_enabled) return;
            const adminFullName = (currentAdmin.full_name && currentAdmin.full_name.trim()) || currentAdmin.username || 'همکار گرامی';
            const confirmed = await showConfirmDialog({
                title: `همکار گرامی، ${adminFullName}`,
                description: 'احراز هویت دومرحله‌ای حساب کاربری شما هنوز فعال نشده است. لطفاً جهت افزایش ضریب امنیت حساب خود، در اسرع وقت نسبت به فعال‌سازی آن اقدام فرمایید.',
                confirmText: 'فعال‌سازی احراز هویت',
                cancelText: 'بعداً یادآوری کن',
                variant: 'warning',
                iconType: 'shield-alert'
            });

            if (confirmed) {
                if (modal2FA) modal2FA.style.display = 'flex';
                reset2FAModalViews();
                await refresh2FAStatus();
                if (btnStartSetup2FA) {
                    btnStartSetup2FA.click();
                }
            }
        }, 400);
    }

    function setupSuperuserUI() {
        const fullnameEl = document.getElementById('header-admin-fullname');
        const roleEl = document.getElementById('header-admin-role');
        const menuFullnameEl = document.getElementById('menu-admin-fullname');
        const menuRoleEl = document.getElementById('menu-admin-role');
        if (currentAdmin) {
            const name = currentAdmin.full_name || currentAdmin.username;
            const role = currentAdmin.role_title || 'کارشناس حراست';
            if (fullnameEl) fullnameEl.textContent = name;
            if (roleEl) roleEl.textContent = role;
            if (menuFullnameEl) menuFullnameEl.textContent = name;
            if (menuRoleEl) menuRoleEl.textContent = role;
        }

        const isSuper = Boolean(currentAdmin && (currentAdmin.is_superuser === true || currentAdmin.role === 'superuser'));
        const deleteReportBtn = document.getElementById('btnDeleteReportSuperuser');
        if (isSuper) {
            if (tabBtnAdvanced) tabBtnAdvanced.style.display = 'inline-flex';
        } else {
            if (tabBtnAdvanced) tabBtnAdvanced.style.display = 'none';
            if (advancedSection) advancedSection.style.display = 'none';
            if (reportsInboxSection) reportsInboxSection.style.display = 'block';
            if (deleteReportBtn) {
                deleteReportBtn.classList.add('d-none');
                deleteReportBtn.style.setProperty('display', 'none', 'important');
            }
        }
    }

    function switchMainTab(tabName) {
        if (tabName === 'reports') {
            tabBtnReports?.classList.add('active');
            tabBtnAnalytics?.classList.remove('active');
            tabBtnAdvanced?.classList.remove('active');
            if (reportsInboxSection) reportsInboxSection.style.display = 'block';
            if (detailSection) detailSection.style.display = 'none';
            if (listSection) listSection.style.display = 'block';
            if (filterContainer) filterContainer.style.display = '';
            reportsInboxSection?.classList.remove('in-detail-view');
            updateBulkToolbar();
            if (advancedSection) advancedSection.style.display = 'none';
            if (analyticsSection) analyticsSection.style.display = 'none';
            fetchReports(true, false).catch(() => { });
            fetchNotifications().catch(() => { });
        } else if (tabName === 'analytics') {
            tabBtnAnalytics?.classList.add('active');
            tabBtnReports?.classList.remove('active');
            tabBtnAdvanced?.classList.remove('active');
            if (reportsInboxSection) reportsInboxSection.style.display = 'none';
            if (detailSection) detailSection.style.display = 'none';
            if (advancedSection) advancedSection.style.display = 'none';
            if (analyticsSection) analyticsSection.style.display = 'block';
            renderAnalyticsDashboard(true).catch(() => { });
        } else if (tabName === 'advanced') {
            if (!currentAdmin || !currentAdmin.is_superuser) {
                showToast('دسترسی به این بخش برای حساب شما مجاز نمی‌باشد.', 3000, 'error');
                return;
            }
            tabBtnAdvanced?.classList.add('active');
            tabBtnReports?.classList.remove('active');
            tabBtnAnalytics?.classList.remove('active');
            if (reportsInboxSection) reportsInboxSection.style.display = 'none';
            if (detailSection) detailSection.style.display = 'none';
            if (analyticsSection) analyticsSection.style.display = 'none';
            if (advancedSection) advancedSection.style.display = 'block';

            const activeSubBtn = document.querySelector('.adv-subnav-btn.active');
            const target = activeSubBtn ? activeSubBtn.getAttribute('data-target') : 'adv-view-users';
            loadAdvSubView(target);
        }
        if (window.TatweelTypography) window.TatweelTypography.recheckOverflow();
    }

    tabBtnReports?.addEventListener('click', () => switchMainTab('reports'));
    tabBtnAnalytics?.addEventListener('click', () => switchMainTab('analytics'));
    tabBtnAdvanced?.addEventListener('click', () => switchMainTab('advanced'));

    document.querySelectorAll('.adv-subnav-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const targetId = btn.getAttribute('data-target');
            document.querySelectorAll('.adv-subnav-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            loadAdvSubView(targetId);
        });
    });

    function loadAdvSubView(targetId) {
        document.querySelectorAll('.adv-subview').forEach(v => {
            if (v.id === targetId) {
                v.style.display = 'block';
            } else {
                v.style.display = 'none';
            }
        });
        if (window.TatweelTypography) window.TatweelTypography.recheckOverflow();

        if (targetId === 'adv-view-users') {
            fetchAdminUsers();
        } else if (targetId === 'adv-view-audit') {
            fetchAuditLogs();
        } else if (targetId === 'adv-view-sessions') {
            fetchActiveSessions();
        } else if (targetId === 'adv-view-system') {
            fetchSystemStatus();
        }
    }

    let allAdminUsers = [];

    function renderUsersTableSkeleton(count = 5) {
        const tbody = document.getElementById('adv-users-tbody');
        if (!tbody) return;
        let html = '';
        for (let i = 0; i < count; i++) {
            html += `
                <tr class="skeleton-tr" aria-hidden="true">
                    <td><div class="skeleton-shimmer" style="width:24px; height:18px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:90px; height:18px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:120px; height:18px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:80px; height:20px; border-radius:10px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:70px; height:18px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:60px; height:20px; border-radius:10px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:110px; height:16px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:70px; height:26px; border-radius:6px;"></div></td>
                </tr>
            `;
        }
        tbody.innerHTML = html;
    }

    async function fetchAdminUsers() {
        const tbody = document.getElementById('adv-users-tbody');
        if (!tbody) return;
        renderUsersTableSkeleton(5);

        try {
            const res = await adminFetch('/api/admin/users');
            if (res.status === 403) {
                tbody.innerHTML = '<tr><td colspan="8" style="text-align:center; padding:20px; color:var(--color-danger);">دسترسی به بخش مدیریت کاربران محدود به پشتیبانی سامانه است.</td></tr>';
                return;
            }
            if (!res.ok) throw new Error('خطا در دریافت کاربران');

            allAdminUsers = await res.json();
            renderUsersTable(allAdminUsers);

            const totalEl = document.getElementById('adv-users-total');
            const activeEl = document.getElementById('adv-users-active');
            const superEl = document.getElementById('adv-users-super');
            if (totalEl) totalEl.textContent = toPersianNum(allAdminUsers.length);
            if (activeEl) activeEl.textContent = toPersianNum(allAdminUsers.filter(u => u.is_active).length);
            if (superEl) superEl.textContent = toPersianNum(allAdminUsers.filter(u => u.is_superuser).length);
        } catch (e) {

            tbody.innerHTML = `<tr><td colspan="8" style="text-align:center; padding:20px; color:var(--color-danger);">خطا: ${escapeHTML(e.message)}</td></tr>`;
        }
    }

    function validatePasswordComplexity(pwd) {
        if (!pwd || pwd.length < 12) {
            return 'رمز عبور باید حداقل ۱۲ کاراکتر باشد.';
        }
        if (!/[A-Z]/.test(pwd)) {
            return 'رمز عبور باید حداقل شامل یک حرف بزرگ انگلیسی (A-Z) باشد.';
        }
        if (!/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~`]/.test(pwd)) {
            return 'رمز عبور باید حداقل شامل یک نماد یا کاراکتر خاص (!@#$%^&*...) باشد.';
        }
        return null;
    }

    function renderUsersTable(users) {
        const tbody = document.getElementById('adv-users-tbody');
        if (!tbody) return;
        tbody.innerHTML = '';

        if (!users || users.length === 0) {
            tbody.innerHTML = '<tr><td colspan="8" style="text-align:center; padding:20px; color:var(--text-muted);">هیچ کاربری یافت نشد.</td></tr>';
            return;
        }

        const fragment = document.createDocumentFragment();

        users.forEach(u => {
            const tr = document.createElement('tr');
            const isSelf = currentAdmin && currentAdmin.username === u.username;

            let lockoutBadge = '';
            if (u.is_locked) {
                lockoutBadge = `<span class="badge-lockout" data-tooltip="قفل موقت ورود به دلیل تلاش‌های ناموفق (باقی‌مانده: ${escapeHTML(u.lockout_remaining || '۱۵ دقیقه')})"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" style="vertical-align:-1px; margin-left:3px;"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>قفل موقت ورود (${escapeHTML(u.lockout_remaining || '')})</span>`;
            }

            const isSuper = u.is_superuser;
            const actionButtons = `
                <button type="button" class="btn-action btn-submit btn-icon-only" data-action="edit-user" data-id="${u.id}" data-tooltip="ویرایش مشخصات مدیر" aria-label="ویرایش مشخصات مدیر">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
                </button>
                <button type="button" class="btn-action btn-secondary btn-icon-only" data-action="reset-pwd" data-id="${u.id}" data-user="${escapeHTML(u.username)}" data-tooltip="تغییر رمز عبور" aria-label="تغییر رمز عبور">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="7.5" cy="15.5" r="5.5"></circle><path d="m21 2-9.6 9.6"></path><path d="m15.5 7.5 2.5 2.5"></path><path d="m18 5 2 2"></path></svg>
                </button>
                <button type="button" class="btn-action ${u.is_locked ? 'btn-warning is-locked-pulse' : 'btn-secondary'} btn-icon-only btn-lift-ban" data-action="lift-ban" data-id="${u.id}" data-user="${escapeHTML(u.username)}" data-tooltip="${u.is_locked ? `رفع قفل موقت ورود (${escapeHTML(u.lockout_remaining || '۱۵ دقیقه')})` : 'رفع محدودیت و بازنشانی خطاهای ورود'}" aria-label="رفع قفل موقت ورود">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 9.9-1"></path></svg>
                </button>
                ${!isSuper ? `
                <button type="button" class="btn-action btn-danger btn-icon-only" data-action="delete-user" data-id="${u.id}" data-user="${escapeHTML(u.username)}" data-tooltip="حذف حساب کاربری مدیر" aria-label="حذف حساب کاربری مدیر">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                </button>
                ` : ''}
            `;

            let pwdBadges = '';
            if (u.must_change_password) {
                pwdBadges += `<span class="badge-pwd-forced" data-tooltip="الزام به تغییر رمز در ورود بعدی"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="vertical-align:-1px; margin-left:3px;"><circle cx="7.5" cy="15.5" r="5.5"></circle><path d="m21 2-9.6 9.6"></path><path d="m15.5 7.5 2.5 2.5"></path><path d="m18 5 2 2"></path></svg>الزام تغییر رمز</span> `;
            }
            if (u.is_password_expired) {
                pwdBadges += `<span class="badge-pwd-expired" data-tooltip="اعتبار ۳۰ روزه منقضی شده است"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="vertical-align:-1px; margin-left:3px;"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>منقضی شده</span> `;
            }

            let roleBadge = '';
            const userRole = u.role || (u.is_superuser ? 'superuser' : 'edare_kol');
            if (userRole === 'superuser') {
                roleBadge = '<span class="badge-super" style="background:rgba(225, 29, 72, 0.12); color:#fda4af; border:1px solid rgba(244, 63, 94, 0.28); font-size:11px; padding:2px 8px; border-radius:12px; font-weight:600;">مدیر ارشد سامانه</span>';
            } else if (userRole === 'navahi_manategh') {
                roleBadge = '<span class="badge-navahi" style="background:rgba(14, 165, 233, 0.15); color:#38bdf8; border:1px solid rgba(56, 189, 248, 0.35); font-size:11px; padding:2px 8px; border-radius:12px; font-weight:600;">نواحی و مناطق</span>';
            } else {
                roleBadge = '<span class="badge-edare-kol" style="background:rgba(99, 102, 241, 0.15); color:#a5b4fc; border:1px solid rgba(165, 180, 252, 0.35); font-size:11px; padding:2px 8px; border-radius:12px; font-weight:600;">اداره کل</span>';
            }

            tr.innerHTML = `
                <td><strong>#${toPersianNum(u.id)}</strong></td>
                <td><strong style="color:var(--text-main); font-weight:600;">${escapeHTML(u.full_name || '-')}</strong></td>
                <td><span class="ltr" style="font-family:Tahoma,sans-serif; color:var(--text-muted);">${escapeHTML(u.username)}</span> ${isSelf ? '<span class="badge badge-read" style="margin-right:6px;">حساب شما</span>' : ''}</td>
                <td>${escapeHTML(u.role_title)}</td>
                <td>${roleBadge}</td>
                <td>
                    ${lockoutBadge ? `<div style="margin-bottom:4px;">${lockoutBadge}</div>` : ''}
                    ${u.is_active ? '<span class="badge-active">فعال</span>' : '<span class="badge-inactive">مسدود / غیرفعال</span>'}
                    ${u.totp_enabled ? '<span class="badge-active" style="display:inline-flex; align-items:center; gap:3px; margin-top:4px;" data-tooltip="احراز هویت دومرحله‌ای فعال است"><svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path><polyline points="9 12 11 14 15 10"></polyline></svg>۲FA فعال</span>' : ''}
                    ${pwdBadges ? `<div style="margin-top:4px;">${pwdBadges}</div>` : ''}
                </td>
                <td class="ltr" style="text-align:right;">${formatJalaliDateTime(u.created_at, false)}</td>
                <td>
                    <div style="display:flex; gap:6px; flex-wrap:wrap; align-items:center;">
                        ${actionButtons}
                    </div>
                </td>
            `;

            fragment.appendChild(tr);
        });

        tbody.appendChild(fragment);

        if (!tbody._delegated) {
            tbody._delegated = true;
            tbody.addEventListener('click', (e) => {
                const b = e.target.closest('button[data-action]');
                if (!b) return;
                const action = b.dataset.action;
                if (action === 'edit-user') {
                    const targetId = parseInt(b.getAttribute('data-id'), 10);
                    const targetUser = allAdminUsers.find(u => u.id === targetId);
                    if (targetUser) openEditUserModal(targetUser);
                } else if (action === 'reset-pwd') {
                    openResetPwdModal(b.getAttribute('data-id'), b.getAttribute('data-user'));
                } else if (action === 'lift-ban') {
                    liftUserLockout(b.getAttribute('data-id'), b.getAttribute('data-user'));
                } else if (action === 'delete-user') {
                    deleteUser(b.getAttribute('data-id'), b.getAttribute('data-user'));
                }
            });
        }
    }

    async function liftUserLockout(userId, username) {
        const confirmed = await showConfirmDialog({
            title: 'رفع قفل موقت ورود',
            description: `آیا از رفع مسدودی موقت و بازنشانی خطاهای ورود حساب کاربری «${username}» اطمینان دارید؟ پس از این اقدام، کاربر می‌تواند مجدداً وارد سامانه شود.`,
            confirmText: 'بله، رفع قفل',
            cancelText: 'انصراف'
        });
        if (!confirmed) return;

        try {
            const res = await adminFetch('/api/admin/users/lift-lockout', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ user_id: parseInt(userId, 10), username: username })
            });
            const data = await res.json();
            if (!res.ok) throw new Error(data.message || 'خطا در رفع قفل کاربر');
            showToast(data.message || 'قفل ورود با موفقیت برداشته شد.', 3000, 'success');
            fetchAdminUsers();
        } catch (e) {
            showToast(e.message, 3000, 'error');
        }
    }

    async function deleteUser(userId, username) {
        const confirmed = await showConfirmDialog({
            title: 'حذف حساب کاربری مدیر',
            description: `آیا از حذف دائم حساب کاربری «${username}» اطمینان دارید؟ این عملیات غیرقابل بازگشت است و دسترسی کاربر فوراً قطع خواهد شد.`,
            confirmText: 'بله، حذف حساب کاربری',
            cancelText: 'انصراف',
            variant: 'danger',
            iconType: 'user-x'
        });
        if (!confirmed) {
            return;
        }

        try {
            const res = await adminFetch('/api/admin/users/delete', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ user_id: parseInt(userId, 10) })
            });
            if (res.ok) {
                showToast(`کاربر "${username}" با موفقیت حذف گردید`);
                fetchAdminUsers();
            } else {
                const err = await res.text();
                showToast(`خطا: ${err}`);
            }
        } catch (e) {
            showToast('خطای ارتباط با سرور.');
        }
    }

    const modalCreateUser = document.getElementById('modal-create-user');
    const modalEditUser = document.getElementById('modal-edit-user');
    const modalResetPassword = document.getElementById('modal-reset-password');
    const btnOpenCreateUser = document.getElementById('btn-open-create-user-modal');
    const btnCloseCreateUser = document.getElementById('btn-close-create-user-modal');
    const btnCancelCreateUser = document.getElementById('btn-cancel-create-user');
    const formCreateUser = document.getElementById('form-create-user');

    const btnCloseEditUser = document.getElementById('btn-close-edit-user-modal');
    const btnCancelEditUser = document.getElementById('btn-cancel-edit-user');
    const formEditUser = document.getElementById('form-edit-user');

    const btnCloseResetPwd = document.getElementById('btn-close-reset-pwd-modal');
    const btnCancelResetPwd = document.getElementById('btn-cancel-reset-pwd');
    const formResetPwd = document.getElementById('form-reset-password');

    const toggleNewPasswordBtn = document.getElementById('toggleNewPasswordBtn');
    const newPasswordInput = document.getElementById('new_password');
    const eyeShowNew = document.getElementById('eye-icon-show-new');
    const eyeHideNew = document.getElementById('eye-icon-hide-new');

    if (toggleNewPasswordBtn && newPasswordInput) {
        toggleNewPasswordBtn.addEventListener('click', () => {
            const isPassword = newPasswordInput.type === 'password';
            newPasswordInput.type = isPassword ? 'text' : 'password';
            if (eyeShowNew && eyeHideNew) {
                eyeShowNew.classList.toggle('d-none', isPassword);
                eyeHideNew.classList.toggle('d-none', !isPassword);
            }
        });
    }

    const toggleNewPasswordConfirmBtn = document.getElementById('toggleNewPasswordConfirmBtn');
    const newPasswordConfirmInput = document.getElementById('new_password_confirm');
    const eyeShowNewConfirm = document.getElementById('eye-icon-show-new-confirm');
    const eyeHideNewConfirm = document.getElementById('eye-icon-hide-new-confirm');

    if (toggleNewPasswordConfirmBtn && newPasswordConfirmInput) {
        toggleNewPasswordConfirmBtn.addEventListener('click', () => {
            const isPassword = newPasswordConfirmInput.type === 'password';
            newPasswordConfirmInput.type = isPassword ? 'text' : 'password';
            if (eyeShowNewConfirm && eyeHideNewConfirm) {
                eyeShowNewConfirm.classList.toggle('d-none', isPassword);
                eyeHideNewConfirm.classList.toggle('d-none', !isPassword);
            }
        });
    }

    function resetNewUserPasswordToggles() {
        if (newPasswordInput) newPasswordInput.type = 'password';
        if (eyeShowNew) eyeShowNew.classList.remove('d-none');
        if (eyeHideNew) eyeHideNew.classList.add('d-none');

        if (newPasswordConfirmInput) newPasswordConfirmInput.type = 'password';
        if (eyeShowNewConfirm) eyeShowNewConfirm.classList.remove('d-none');
        if (eyeHideNewConfirm) eyeHideNewConfirm.classList.add('d-none');
    }

    btnOpenCreateUser?.addEventListener('click', () => {
        formCreateUser?.reset();
        resetNewUserPasswordToggles();
        syncAllAdminCustomSelects();
        const status = document.getElementById('create-user-status');
        if (status) status.style.display = 'none';
        if (modalCreateUser) modalCreateUser.style.display = 'flex';
    });

    const closeCreateUserModal = () => {
        if (modalCreateUser) modalCreateUser.style.display = 'none';
        resetNewUserPasswordToggles();
    };
    btnCloseCreateUser?.addEventListener('click', closeCreateUserModal);
    btnCancelCreateUser?.addEventListener('click', closeCreateUserModal);

    function openEditUserModal(user) {
        if (!user) return;
        formEditUser?.reset();
        const idInput = document.getElementById('edit_user_id');
        const userInput = document.getElementById('edit_username');
        const nameInput = document.getElementById('edit_full_name');
        const roleInput = document.getElementById('edit_role_title');
        const roleSelect = document.getElementById('edit_role');
        const activeSelect = document.getElementById('edit_is_active');
        const activeContainer = document.getElementById('edit_active_container');
        const forcePwdCheckbox = document.getElementById('edit_must_change_pwd');
        const statusEl = document.getElementById('edit-user-status');

        if (idInput) idInput.value = user.id;
        if (userInput) userInput.value = user.username;
        if (nameInput) nameInput.value = user.full_name || '';
        if (roleInput) roleInput.value = user.role_title || '';
        if (roleSelect) {
            roleSelect.value = user.role || (user.is_superuser ? 'superuser' : 'edare_kol');
        }
        if (forcePwdCheckbox) forcePwdCheckbox.checked = !!user.must_change_password;

        if (user.is_superuser) {
            if (activeSelect) activeSelect.value = 'true';
            if (activeContainer) activeContainer.style.display = 'none';
        } else {
            if (activeContainer) activeContainer.style.display = 'block';
            if (activeSelect) activeSelect.value = user.is_active ? 'true' : 'false';
        }
        syncAllAdminCustomSelects();

        if (statusEl) statusEl.style.display = 'none';

        const lockoutBanner = document.getElementById('edit-lockout-banner');
        const lockoutText = document.getElementById('edit-lockout-text');
        const btnModalLift = document.getElementById('btn-modal-lift-lockout');
        if (lockoutBanner) {
            if (user.is_locked) {
                lockoutBanner.style.display = 'flex';
                if (lockoutText) lockoutText.textContent = `حساب کاربری به دلیل تلاش‌های ناموفق ورود مسدود است (${user.lockout_remaining || '۱۵ دقیقه'}).`;
                if (btnModalLift) {
                    btnModalLift.onclick = async () => {
                        await liftUserLockout(user.id, user.username);
                        user.is_locked = false;
                        lockoutBanner.style.display = 'none';
                    };
                }
            } else {
                lockoutBanner.style.display = 'none';
            }
        }

        const iconContainer = document.getElementById('edit_2fa_icon_container');
        const iconActive = document.getElementById('edit_2fa_icon_active');
        const iconInactive = document.getElementById('edit_2fa_icon_inactive');
        const badge2FA = document.getElementById('edit_2fa_badge');
        const text2FA = document.getElementById('edit_2fa_status_text');
        const disableBtn2FA = document.getElementById('btn-edit-disable-2fa');

        const update2FAState = (has2FA) => {
            if (has2FA) {
                if (iconContainer) {
                    iconContainer.style.background = 'rgba(16, 185, 129, 0.15)';
                    iconContainer.style.color = '#34d399';
                }
                if (iconActive) iconActive.style.display = 'block';
                if (iconInactive) iconInactive.style.display = 'none';
                if (badge2FA) badge2FA.className = 'badge-2fa-has';
                if (text2FA) text2FA.textContent = 'این کاربر احراز هویت دو مرحله ای دارد';
                if (disableBtn2FA) {
                    disableBtn2FA.style.display = 'inline-flex';
                    disableBtn2FA.disabled = false;
                }
            } else {
                if (iconContainer) {
                    iconContainer.style.background = 'rgba(244, 63, 94, 0.15)';
                    iconContainer.style.color = '#f43f5e';
                }
                if (iconActive) iconActive.style.display = 'none';
                if (iconInactive) iconInactive.style.display = 'block';
                if (badge2FA) badge2FA.className = 'badge-2fa-none';
                if (text2FA) text2FA.textContent = 'این کاربر احراز هویت دو مرحله ای ندارد';
                if (disableBtn2FA) disableBtn2FA.style.display = 'none';
            }
        };

        update2FAState(!!user.totp_enabled);

        if (disableBtn2FA) {
            disableBtn2FA.onclick = async () => {
                const confirmed = await showConfirmDialog({
                    title: 'غیرفعال‌سازی احراز هویت دومرحله‌ای',
                    description: `آیا از غیرفعال‌سازی احراز هویت دومرحله‌ای حساب «${user.username}» اطمینان دارید؟`,
                    confirmText: 'بله، غیرفعال کن',
                    cancelText: 'انصراف',
                    variant: 'danger',
                    iconType: 'shield-alert'
                });
                if (!confirmed) return;

                disableBtn2FA.disabled = true;
                try {
                    const res = await adminFetch('/api/admin/users/disable-2fa', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ user_id: user.id, username: user.username })
                    });
                    const data = await res.json();
                    if (!res.ok) throw new Error(data.message || 'خطا در غیرفعال‌سازی احراز هویت دومرحله‌ای');

                    user.totp_enabled = false;
                    const match = allAdminUsers.find(u => u.id === user.id);
                    if (match) match.totp_enabled = false;

                    update2FAState(false);
                    showToast(data.message || 'احراز هویت دومرحله‌ای با موفقیت غیرفعال شد.', 3500, 'success');
                    renderUsersTable(allAdminUsers);

                    if (currentAdmin && currentAdmin.id === user.id) {
                        currentAdmin2FAEnabled = false;
                        const statusBadge = document.getElementById('current-admin-2fa-status');
                        if (statusBadge) statusBadge.textContent = 'غیرفعال';
                    }
                } catch (err) {
                    showToast(err.message || 'خطا در غیرفعال‌سازی ۲FA', 3500, 'error');
                    disableBtn2FA.disabled = false;
                }
            };
        }

        if (modalEditUser) {

            const card = modalEditUser.querySelector('.modal-card') || modalEditUser;
            promoteTemporarily(card, 'transform, opacity', 250);
            modalEditUser.style.display = 'flex';
        }
    }

    const closeEditUserModal = () => { if (modalEditUser) modalEditUser.style.display = 'none'; };
    btnCloseEditUser?.addEventListener('click', closeEditUserModal);
    btnCancelEditUser?.addEventListener('click', closeEditUserModal);

    formEditUser?.addEventListener('submit', async (e) => {
        e.preventDefault();
        const submitBtn = document.getElementById('btn-submit-edit-user');
        const btnText = submitBtn?.querySelector('.btn-text');
        const btnSpinner = submitBtn?.querySelector('.btn-spinner');
        const statusEl = document.getElementById('edit-user-status');

        const userId = parseInt(document.getElementById('edit_user_id')?.value, 10);
        const fullName = document.getElementById('edit_full_name')?.value.trim();
        const roleTitle = document.getElementById('edit_role_title')?.value.trim();
        const role = document.getElementById('edit_role')?.value || 'edare_kol';
        const isActive = document.getElementById('edit_is_active')?.value === 'true';
        const mustChangePwd = document.getElementById('edit_must_change_pwd')?.checked || false;

        if (submitBtn) submitBtn.disabled = true;
        if (btnText) btnText.textContent = 'در حال ذخیره...';
        if (btnSpinner) btnSpinner.classList.remove('d-none');
        if (statusEl) statusEl.style.display = 'none';

        try {
            const res = await adminFetch('/api/admin/users/update', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    user_id: userId,
                    full_name: fullName,
                    role_title: roleTitle,
                    role: role,
                    is_active: isActive,
                    must_change_password: mustChangePwd
                })
            });

            if (res.ok) {
                showToast(`اطلاعات کاربر "${fullName}" با موفقیت بروزرسانی شد`);
                closeEditUserModal();
                fetchAdminUsers();
                if (currentAdmin && currentAdmin.id === userId) {
                    await checkAdminProfile();
                }
            } else {
                const err = await res.text();
                if (statusEl) {
                    statusEl.textContent = err || 'خطا در ویرایش کاربر.';
                    statusEl.style.display = 'block';
                }
            }
        } catch (err) {
            if (statusEl) {
                statusEl.textContent = 'خطای ارتباط با سرور.';
                statusEl.style.display = 'block';
            }
        } finally {
            if (submitBtn) submitBtn.disabled = false;
            if (btnText) btnText.textContent = 'ذخیره تغییرات';
            if (btnSpinner) btnSpinner.classList.add('d-none');
        }
    });

    formCreateUser?.addEventListener('submit', async (e) => {
        e.preventDefault();
        const submitBtn = document.getElementById('btn-submit-create-user');
        const btnText = submitBtn?.querySelector('.btn-text');
        const btnSpinner = submitBtn?.querySelector('.btn-spinner');
        const statusEl = document.getElementById('create-user-status');

        const passwordVal = document.getElementById('new_password')?.value.trim();
        const passwordConfirmVal = document.getElementById('new_password_confirm')?.value.trim();

        if (passwordVal !== passwordConfirmVal) {
            if (statusEl) {
                statusEl.textContent = 'کلمه عبور با تکرار آن مطابقت ندارد.';
                statusEl.style.display = 'block';
            }
            return;
        }

        const pwdError = validatePasswordComplexity(passwordVal);
        if (pwdError) {
            if (statusEl) {
                statusEl.textContent = pwdError;
                statusEl.style.display = 'block';
            }
            return;
        }

        if (submitBtn) submitBtn.disabled = true;
        if (btnText) btnText.textContent = 'در حال ثبت...';
        if (btnSpinner) btnSpinner.classList.remove('d-none');
        if (statusEl) statusEl.style.display = 'none';

        const payload = {
            full_name: document.getElementById('new_full_name')?.value.trim(),
            username: document.getElementById('new_username')?.value.trim(),
            password: passwordVal,
            role_title: document.getElementById('new_role_title')?.value.trim(),
            role: document.getElementById('new_role')?.value || 'edare_kol',
            must_change_password: document.getElementById('create_must_change_pwd')?.checked || false
        };

        try {
            const res = await adminFetch('/api/admin/users/create', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (res.ok) {
                showToast(`کاربر "${payload.full_name || payload.username}" با موفقیت ثبت شد`);
                closeCreateUserModal();
                fetchAdminUsers();
            } else {
                const err = await res.text();
                if (statusEl) {
                    statusEl.textContent = err || 'خطا در ثبت کاربر.';
                    statusEl.style.display = 'block';
                }
            }
        } catch (err) {
            if (statusEl) {
                statusEl.textContent = 'خطای ارتباط با سرور.';
                statusEl.style.display = 'block';
            }
        } finally {
            if (submitBtn) submitBtn.disabled = false;
            if (btnText) btnText.textContent = 'ثبت کاربر';
            if (btnSpinner) btnSpinner.classList.add('d-none');
        }
    });

    function openResetPwdModal(userId, username) {
        formResetPwd?.reset();
        const idInput = document.getElementById('reset_pwd_user_id');
        const labelEl = document.getElementById('reset_pwd_username_label');
        const statusEl = document.getElementById('reset-pwd-status');

        if (idInput) idInput.value = userId;
        if (labelEl) labelEl.textContent = username;
        if (statusEl) statusEl.style.display = 'none';
        if (modalResetPassword) {

            const card = modalResetPassword.querySelector('.modal-card') || modalResetPassword;
            promoteTemporarily(card, 'transform, opacity', 250);
            modalResetPassword.style.display = 'flex';
        }
    }

    const closeResetPwdModal = () => {
        if (modalResetPassword) {
            modalResetPassword.style.display = 'none';
            resetPasswordWrappers(modalResetPassword);
        }
    };
    btnCloseResetPwd?.addEventListener('click', closeResetPwdModal);
    btnCancelResetPwd?.addEventListener('click', closeResetPwdModal);

    formResetPwd?.addEventListener('submit', async (e) => {
        e.preventDefault();
        const submitBtn = document.getElementById('btn-submit-reset-pwd');
        const btnText = submitBtn?.querySelector('.btn-text');
        const btnSpinner = submitBtn?.querySelector('.btn-spinner');
        const statusEl = document.getElementById('reset-pwd-status');

        const userId = parseInt(document.getElementById('reset_pwd_user_id')?.value, 10);
        const newPassword = document.getElementById('reset_new_password')?.value.trim();
        const mustChangePwd = document.getElementById('reset_must_change_pwd')?.checked || false;

        const pwdError = validatePasswordComplexity(newPassword);
        if (pwdError) {
            if (statusEl) {
                statusEl.textContent = pwdError;
                statusEl.style.display = 'block';
            }
            return;
        }

        if (submitBtn) submitBtn.disabled = true;
        if (btnText) btnText.textContent = 'در حال تغییر...';
        if (btnSpinner) btnSpinner.classList.remove('d-none');
        if (statusEl) statusEl.style.display = 'none';

        try {
            const res = await adminFetch('/api/admin/users/reset-password', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    user_id: userId,
                    new_password: newPassword,
                    must_change_password: mustChangePwd
                })
            });

            if (res.ok) {
                showToast('رمز عبور با موفقیت بروزرسانی شد');
                closeResetPwdModal();
                fetchAdminUsers();
            } else {
                const err = await res.text();
                if (statusEl) {
                    statusEl.textContent = err || 'خطا در تغییر پسورد.';
                    statusEl.style.display = 'block';
                }
            }
        } catch (err) {
            if (statusEl) {
                statusEl.textContent = 'خطای ارتباط با سرور.';
                statusEl.style.display = 'block';
            }
        } finally {
            if (submitBtn) submitBtn.disabled = false;
            if (btnText) btnText.textContent = 'ثبت رمز جدید';
            if (btnSpinner) btnSpinner.classList.add('d-none');
        }
    });

    const modalChangeMyPwd = document.getElementById('modal-change-my-password');
    const btnOpenChangeMyPwd = document.getElementById('btnOpenChangeMyPwd');
    const btnCloseChangeMyPwd = document.getElementById('btn-close-change-my-pwd-modal');
    const btnCancelChangeMyPwd = document.getElementById('btn-cancel-change-my-pwd');
    const formChangeMyPwd = document.getElementById('form-change-my-password');

    btnOpenChangeMyPwd?.addEventListener('click', () => {
        formChangeMyPwd?.reset();
        const status = document.getElementById('change-my-pwd-status');
        if (status) status.style.display = 'none';
        if (modalChangeMyPwd) {

            const card = modalChangeMyPwd.querySelector('.modal-card') || modalChangeMyPwd;
            promoteTemporarily(card, 'transform, opacity', 250);
            modalChangeMyPwd.style.display = 'flex';
        }
    });

    const closeChangeMyPwdModal = () => {
        if (modalChangeMyPwd) {
            modalChangeMyPwd.style.display = 'none';
            resetPasswordWrappers(modalChangeMyPwd);
        }
    };
    btnCloseChangeMyPwd?.addEventListener('click', closeChangeMyPwdModal);
    btnCancelChangeMyPwd?.addEventListener('click', closeChangeMyPwdModal);

    formChangeMyPwd?.addEventListener('submit', async (e) => {
        e.preventDefault();
        const curPwd = document.getElementById('my_current_password')?.value.trim();
        const newPwd = document.getElementById('my_new_password')?.value.trim();
        const confPwd = document.getElementById('my_confirm_password')?.value.trim();
        const statusEl = document.getElementById('change-my-pwd-status');
        const submitBtn = document.getElementById('btn-submit-change-my-pwd');
        const btnText = submitBtn?.querySelector('.btn-text');
        const btnSpinner = submitBtn?.querySelector('.btn-spinner');

        if (newPwd !== confPwd) {
            if (statusEl) {
                statusEl.textContent = 'کلمه عبور جدید با تکرار آن مطابقت ندارد.';
                statusEl.style.display = 'block';
            }
            return;
        }

        const pwdError = validatePasswordComplexity(newPwd);
        if (pwdError) {
            if (statusEl) {
                statusEl.textContent = pwdError;
                statusEl.style.display = 'block';
            }
            return;
        }

        if (submitBtn) submitBtn.disabled = true;
        if (btnText) btnText.textContent = 'در حال ذخیره...';
        if (btnSpinner) btnSpinner.classList.remove('d-none');
        if (statusEl) statusEl.style.display = 'none';

        try {
            const res = await adminFetch('/api/admin/change-my-password', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ current_password: curPwd, new_password: newPwd })
            });

            if (res.ok) {
                showToast('کلمه عبور شما با موفقیت تغییر یافت');
                closeChangeMyPwdModal();
                if (currentAdmin) {
                    currentAdmin.must_change_password = false;
                    currentAdmin.is_password_expired = false;
                }
            } else {
                const err = await res.text();
                if (statusEl) {
                    statusEl.textContent = err || 'خطا در تغییر رمز عبور.';
                    statusEl.style.display = 'block';
                }
            }
        } catch (err) {
            if (statusEl) {
                statusEl.textContent = 'خطای ارتباط با سرور.';
                statusEl.style.display = 'block';
            }
        } finally {
            if (submitBtn) submitBtn.disabled = false;
            if (btnText) btnText.textContent = 'تغییر کلمه عبور';
            if (btnSpinner) btnSpinner.classList.add('d-none');
        }
    });

    const modalForceChangePwd = document.getElementById('modal-force-change-password');
    const formForceChangePwd = document.getElementById('form-force-change-password');
    const btnForcePwdLogout = document.getElementById('btn-force-pwd-logout');

    btnForcePwdLogout?.addEventListener('click', async () => {
        try {
            await adminFetch('/api/admin/logout', { method: 'POST' });
        } catch (e) { }
        document.cookie = 'herasat_admin_session=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
        location.reload();
    });

    formForceChangePwd?.addEventListener('submit', async (e) => {
        e.preventDefault();
        const curPwd = document.getElementById('force_current_password')?.value.trim();
        const newPwd = document.getElementById('force_new_password')?.value.trim();
        const confPwd = document.getElementById('force_confirm_password')?.value.trim();
        const statusEl = document.getElementById('force-change-pwd-status');
        const submitBtn = document.getElementById('btn-submit-force-change-pwd');
        const btnText = submitBtn?.querySelector('.btn-text');
        const btnSpinner = submitBtn?.querySelector('.btn-spinner');

        if (newPwd !== confPwd) {
            if (statusEl) {
                statusEl.textContent = 'کلمه عبور جدید با تکرار آن مطابقت ندارد.';
                statusEl.style.display = 'block';
            }
            return;
        }

        const pwdError = validatePasswordComplexity(newPwd);
        if (pwdError) {
            if (statusEl) {
                statusEl.textContent = pwdError;
                statusEl.style.display = 'block';
            }
            return;
        }

        if (submitBtn) submitBtn.disabled = true;
        if (btnText) btnText.textContent = 'در حال ثبت...';
        if (btnSpinner) btnSpinner.classList.remove('d-none');
        if (statusEl) statusEl.style.display = 'none';

        try {
            const res = await adminFetch('/api/admin/change-my-password', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ current_password: curPwd, new_password: newPwd })
            });

            if (res.ok) {
                showToast('کلمه عبور با موفقیت تغییر یافت و ثبت گردید');
                if (modalForceChangePwd) {
                    modalForceChangePwd.style.display = 'none';
                    resetPasswordWrappers(modalForceChangePwd);
                }
                if (currentAdmin) {
                    currentAdmin.must_change_password = false;
                    currentAdmin.is_password_expired = false;
                    check2FAEnforcementNotice();
                }
            } else {
                const err = await res.text();
                if (statusEl) {
                    statusEl.textContent = err || 'خطا در تغییر رمز عبور.';
                    statusEl.style.display = 'block';
                }
            }
        } catch (err) {
            if (statusEl) {
                statusEl.textContent = 'خطای ارتباط با سرور.';
                statusEl.style.display = 'block';
            }
        } finally {
            if (submitBtn) submitBtn.disabled = false;
            if (btnText) btnText.textContent = 'ثبت کلمه عبور جدید و ورود';
            if (btnSpinner) btnSpinner.classList.add('d-none');
        }
    });

    function resetPasswordWrappers(container) {
        if (!container) return;
        container.querySelectorAll('.password-wrapper').forEach(wrapper => {
            const input = wrapper.querySelector('input');
            if (input) input.type = 'password';
            const eyeShow = wrapper.querySelector('.eye-icon-show');
            const eyeHide = wrapper.querySelector('.eye-icon-hide');
            if (eyeShow) eyeShow.classList.remove('d-none');
            if (eyeHide) eyeHide.classList.add('d-none');
        });
    }

    document.addEventListener('click', (e) => {
        const btn = e.target.closest('.password-wrapper .btn-toggle-pw');
        if (!btn) return;
        if (btn.id === 'toggleNewPasswordBtn' || btn.id === 'toggleNewPasswordConfirmBtn' || btn.id === 'togglePasswordBtn') return;
        e.preventDefault();
        const wrapper = btn.closest('.password-wrapper');
        if (!wrapper) return;
        const input = wrapper.querySelector('input');
        if (!input) return;
        const isPassword = input.type === 'password';
        input.type = isPassword ? 'text' : 'password';
        const eyeShow = btn.querySelector('.eye-icon-show');
        const eyeHide = btn.querySelector('.eye-icon-hide');
        if (eyeShow && eyeHide) {
            eyeShow.classList.toggle('d-none', isPassword);
            eyeHide.classList.toggle('d-none', !isPassword);
        }
    });

    const modalConcurrentAlert = document.getElementById('modal-concurrent-alert');
    const btnAckConcurrentLogout = document.getElementById('btn-ack-concurrent-logout');

    let isKickoutHandled = false;

    function triggerConcurrentLoginAlert() {
        if (isKickoutHandled) return;
        isKickoutHandled = true;
        clearInterval(timerInterval);

        if (_sseConnection) {
            _sseConnection.close();
            _sseConnection = null;
        }

        if (modalConcurrentAlert) {
            modalConcurrentAlert.style.display = 'flex';
        } else {
            showConfirmDialog({
                title: 'خروج همزمان از حساب کاربری',
                description: 'یک ورود جدید با نام کاربری شما در دستگاه دیگری انجام شد و شما از این سامانه خارج شدید.',
                confirmText: 'ورود مجدد به سامانه',
                cancelText: 'بستن',
                variant: 'danger',
                iconType: 'shield-alert'
            }).then(() => {
                location.reload();
            });
        }
    }

    btnAckConcurrentLogout?.addEventListener('click', () => {
        sessionStorage.removeItem('2fa_prompt_shown');
        document.cookie = 'herasat_admin_session=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
        location.reload();
    });

    let allAuditLogs = [];
    let filteredAuditLogs = [];
    let auditCurrentPage = 1;
    const auditPageSize = 25;

    const auditFilterSearch = document.getElementById('audit-filter-search');
    const auditFilterAction = document.getElementById('audit-filter-action');
    const auditFilterUser = document.getElementById('audit-filter-user');
    const auditClearFiltersBtn = document.getElementById('audit-clear-filters');
    const btnRefreshAudit = document.getElementById('btn-refresh-audit');

    function populateAuditUsersDropdown(logs) {
        if (!auditFilterUser) return;
        const currentVal = auditFilterUser.value;
        const users = Array.from(new Set(logs.map(l => l.admin_username).filter(Boolean))).sort();

        let html = '<option value="all">همه کاربران اداری</option>';
        users.forEach(u => {
            html += `<option value="${escapeHTML(u)}">${escapeHTML(u)}</option>`;
        });
        auditFilterUser.innerHTML = html;
        if (users.includes(currentVal)) {
            auditFilterUser.value = currentVal;
        }
        if (typeof initAdminCustomSelect === 'function') {
            initAdminCustomSelect('audit-filter-user', ADMIN_FILTER_DESCRIPTIONS['audit-filter-user']);
        }
    }

    function applyAuditFiltersAndPagination(resetPage = false) {
        if (resetPage) auditCurrentPage = 1;

        const searchQuery = auditFilterSearch?.value.trim().toLowerCase() || '';
        const actionQuery = auditFilterAction?.value || 'all';
        const userQuery = auditFilterUser?.value || 'all';

        filteredAuditLogs = allAuditLogs.filter(l => {

            if (searchQuery) {
                const inIp = (l.ip_address || '').toLowerCase().includes(searchQuery);
                const inUser = (l.admin_username || '').toLowerCase().includes(searchQuery);
                const inTitle = (l.admin_title || '').toLowerCase().includes(searchQuery);
                const inAction = (l.action || '').toLowerCase().includes(searchQuery);
                const inTarget = l.target_report_id ? String(l.target_report_id).includes(searchQuery) : false;
                const inTracking = (l.target_tracking_code || '').toLowerCase().includes(searchQuery);
                if (!inIp && !inUser && !inTitle && !inAction && !inTarget && !inTracking) return false;
            }

            if (actionQuery !== 'all') {
                if (l.action !== actionQuery) return false;
            }

            if (userQuery !== 'all') {
                if (l.admin_username !== userQuery) return false;
            }

            if (auditActiveShamsiDateRange && auditActiveShamsiDateRange.startKey && auditActiveShamsiDateRange.endKey) {
                const lKey = l._jalaliDateKey || (l.created_at ? getJalaliDateKeyFromDate(parseTimestamp(l.created_at)) : 0);
                if (lKey < auditActiveShamsiDateRange.startKey || lKey > auditActiveShamsiDateRange.endKey) return false;
            } else if (auditActiveShamsiDateRange && auditActiveShamsiDateRange.startKey) {
                const lKey = l._jalaliDateKey || (l.created_at ? getJalaliDateKeyFromDate(parseTimestamp(l.created_at)) : 0);
                if (lKey < auditActiveShamsiDateRange.startKey) return false;
            }

            return true;
        });

        const totalItems = filteredAuditLogs.length;
        const totalPages = Math.ceil(totalItems / auditPageSize) || 1;
        if (auditCurrentPage > totalPages) auditCurrentPage = totalPages;
        if (auditCurrentPage < 1) auditCurrentPage = 1;

        const startIdx = (auditCurrentPage - 1) * auditPageSize;
        const endIdx = startIdx + auditPageSize;
        const pageItems = filteredAuditLogs.slice(startIdx, endIdx);

        renderAuditTable(pageItems);

        const paginationInfo = document.getElementById('audit-pagination-info');
        const pageBadge = document.getElementById('audit-current-page-badge');
        const btnFirst = document.getElementById('btn-audit-first');
        const btnPrev = document.getElementById('btn-audit-prev');
        const btnNext = document.getElementById('btn-audit-next');
        const btnLast = document.getElementById('btn-audit-last');

        if (paginationInfo) {
            if (totalItems === 0) {
                paginationInfo.textContent = 'لاگ ۰ از ۰';
            } else {
                const fromNum = startIdx + 1;
                const toNum = Math.min(endIdx, totalItems);
                paginationInfo.textContent = `نمایش ${toPersianNum(fromNum)} تا ${toPersianNum(toNum)} از ${toPersianNum(totalItems)} لاگ`;
            }
        }

        if (pageBadge) {
            pageBadge.textContent = `${toPersianNum(auditCurrentPage)} / ${toPersianNum(totalPages)}`;
        }

        if (btnFirst) btnFirst.disabled = (auditCurrentPage === 1 || totalPages <= 1);
        if (btnPrev) btnPrev.disabled = (auditCurrentPage === 1 || totalPages <= 1);
        if (btnNext) btnNext.disabled = (auditCurrentPage === totalPages || totalPages <= 1);
        if (btnLast) btnLast.disabled = (auditCurrentPage === totalPages || totalPages <= 1);
    }

    function initAuditActionSearchBar() {
        const wrapper = document.getElementById('auditActionSearchWrapper');
        const input = document.getElementById('audit-filter-search');
        const dropdown = document.getElementById('auditActionSearchDropdown');
        const list = document.getElementById('auditActionSearchList');
        const titleEl = document.getElementById('auditActionSearchDropdownTitle');
        const countEl = document.getElementById('auditActionSearchCount');
        const iconSearch = document.getElementById('auditActionIconSearch');
        const iconSend = document.getElementById('auditActionIconSend');
        const iconBox = document.getElementById('auditSearchIconBox');
        const kbdBadge = document.getElementById('auditSearchKbdBadge');
        const toggleBtn = document.getElementById('auditExpandableSearchToggle');
        const clearBtn = document.getElementById('auditExpandableSearchClear');
        const iconSearchToggle = toggleBtn ? toggleBtn.querySelector('.icon-search') : null;
        const iconCloseToggle = toggleBtn ? toggleBtn.querySelector('.icon-close') : null;

        if (!wrapper || !input || !dropdown || !list) return;

        let activeIndex = -1;
        let currentItems = [];

        function openSearch() {
            wrapper.classList.add('is-open');
            if (iconSearchToggle) iconSearchToggle.classList.add('d-none');
            if (iconCloseToggle) iconCloseToggle.classList.remove('d-none');
            if (toggleBtn) {
                toggleBtn.setAttribute('aria-label', 'بستن جستجو');
                toggleBtn.setAttribute('data-tooltip', 'بستن جستجو (Esc)');
            }
            if (input.value.trim().length > 0 && clearBtn) {
                clearBtn.classList.remove('d-none');
            }
            setTimeout(() => {
                input.focus();
            }, 60);
        }

        function closeSearch(clearValue = true) {
            if (clearValue && input.value) {
                input.value = '';
                if (typeof applyAuditFiltersAndPagination === 'function') {
                    applyAuditFiltersAndPagination(true);
                }
            }
            wrapper.classList.remove('is-open');
            if (iconSearchToggle) iconSearchToggle.classList.remove('d-none');
            if (iconCloseToggle) iconCloseToggle.classList.add('d-none');
            if (toggleBtn) {
                toggleBtn.setAttribute('aria-label', 'جستجو در لاگ‌ها');
                toggleBtn.setAttribute('data-tooltip', 'جستجو در لاگ‌های امنیتی');
            }
            if (clearBtn) clearBtn.classList.add('d-none');
            if (kbdBadge) kbdBadge.style.opacity = '1';
            closeDropdown();
            input.blur();
        }

        if (toggleBtn) {
            toggleBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                if (wrapper.classList.contains('is-open')) {
                    if (input.value.trim().length > 0) {
                        closeSearch(true);
                    } else {
                        closeSearch(false);
                    }
                } else {
                    openSearch();
                    openDropdown();
                }
            });
        }

        if (clearBtn) {
            clearBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                input.value = '';
                clearBtn.classList.add('d-none');
                if (kbdBadge) kbdBadge.style.opacity = '1';
                input.focus();
                computeAndRender();
                if (typeof applyAuditFiltersAndPagination === 'function') {
                    applyAuditFiltersAndPagination(true);
                }
            });
        }

        const defaultActions = [
            {
                id: 'act-audit-login',
                title: 'ورود کاربران اداری (LOGIN)',
                desc: 'فیلتر سریع رویدادهای ورود به سامانه نظارتی',
                tag: 'احراز هویت',
                iconBoxClass: 'action-icon-blue',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"></path><polyline points="10 17 15 12 10 7"></polyline><line x1="15" y1="12" x2="3" y2="12"></line></svg>',
                run: () => {
                    if (auditFilterAction) auditFilterAction.value = 'LOGIN';
                    syncAllAdminCustomSelects();
                    applyAuditFiltersAndPagination(true);
                }
            },
            {
                id: 'act-audit-logout',
                title: 'خروج کاربران اداری (LOGOUT)',
                desc: 'فیلتر رویدادهای خروج رسمی از حساب کاربری',
                tag: 'نشست',
                iconBoxClass: 'action-icon-slate',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>',
                run: () => {
                    if (auditFilterAction) auditFilterAction.value = 'LOGOUT';
                    syncAllAdminCustomSelects();
                    applyAuditFiltersAndPagination(true);
                }
            },
            {
                id: 'act-audit-reports',
                title: 'بررسی و پاسخ پرونده‌ها',
                desc: 'فیلتر مشاهده پرونده و ثبت پاسخ‌های نظارتی',
                tag: 'پرونده‌ها',
                iconBoxClass: 'action-icon-amber',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line></svg>',
                run: () => {
                    if (auditFilterAction) auditFilterAction.value = 'RESPOND_REPORT';
                    syncAllAdminCustomSelects();
                    applyAuditFiltersAndPagination(true);
                }
            },
            {
                id: 'act-audit-users',
                title: 'مدیریت و دسترسی کاربران',
                desc: 'فیلتر ایجاد، حذف، تغییر رمز و فعال‌سازی حساب',
                tag: 'کاربران',
                iconBoxClass: 'action-icon-purple',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>',
                run: () => {
                    if (auditFilterAction) auditFilterAction.value = 'CREATE_ADMIN_USER';
                    syncAllAdminCustomSelects();
                    applyAuditFiltersAndPagination(true);
                }
            },
            {
                id: 'act-audit-revokes',
                title: 'ابطال نشست‌های امنیتی',
                desc: 'فیلتر لغو نشست‌های دستگاه و نشست‌های فعال دیگر',
                tag: 'امنیت',
                iconBoxClass: 'action-icon-rose',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>',
                run: () => {
                    if (auditFilterAction) auditFilterAction.value = 'REVOKE_SESSION';
                    syncAllAdminCustomSelects();
                    applyAuditFiltersAndPagination(true);
                }
            },
            {
                id: 'act-audit-clear',
                title: 'بازنشانی فیلترها و لاگ‌ها',
                desc: 'حذف تمامی فیلترها و نمایش کامل لاگ‌های امنیتی',
                tag: 'بازنشانی',
                iconBoxClass: 'action-icon-blue',
                iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"></polyline><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"></path></svg>',
                run: () => {
                    if (auditClearFiltersBtn) auditClearFiltersBtn.click();
                }
            }
        ];

        function normalizeQuery(str) {
            if (!str) return '';
            const fa2en = { '۰': '0', '۱': '1', '۲': '2', '۳': '3', '۴': '4', '۵': '5', '۶': '6', '۷': '7', '۸': '8', '۹': '9' };
            const ar2en = { '٠': '0', '١': '1', '٢': '2', '٣': '3', '٤': '4', '٥': '5', '٦': '6', '٧': '7', '٨': '8', '٩': '9' };
            return String(str).replace(/[۰-۹]/g, c => fa2en[c]).replace(/[٠-٩]/g, c => ar2en[c]).toLowerCase().trim();
        }

        function renderItems(items, title) {
            currentItems = items;
            activeIndex = -1;
            list.innerHTML = '';
            if (titleEl) titleEl.textContent = title;
            if (countEl) countEl.textContent = toPersianDigits(items.length);

            if (items.length === 0) {
                const emptyDiv = document.createElement('div');
                emptyDiv.className = 'action-search-empty';
                emptyDiv.innerHTML = '<svg class="action-empty-icon" viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line><line x1="8" y1="11" x2="14" y2="11"></line></svg><span>رویدادی منطبق با جستجو یافت نشد</span>';
                list.appendChild(emptyDiv);
                return;
            }

            items.forEach((item, idx) => {
                const row = document.createElement('div');
                row.className = 'action-search-item';
                row.setAttribute('role', 'option');
                row.setAttribute('id', `audit-action-item-${idx}`);
                row.dataset.index = String(idx);

                const mainCol = document.createElement('div');
                mainCol.className = 'action-search-item-main';

                const iconBox = document.createElement('span');
                iconBox.className = `action-item-icon-box ${item.iconBoxClass || 'action-icon-blue'}`;
                iconBox.innerHTML = item.iconSvg || '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle></svg>';

                const texts = document.createElement('div');
                texts.className = 'action-item-texts';

                const title = document.createElement('span');
                title.className = 'action-item-title';
                title.textContent = item.title;

                const desc = document.createElement('span');
                desc.className = 'action-item-desc';
                desc.textContent = item.desc;

                texts.appendChild(title);
                texts.appendChild(desc);
                mainCol.appendChild(iconBox);
                mainCol.appendChild(texts);

                const meta = document.createElement('div');
                meta.className = 'action-item-meta';

                if (item.short) {
                    const kbd = document.createElement('kbd');
                    kbd.className = 'action-kbd-shortcut';
                    kbd.textContent = item.short;
                    meta.appendChild(kbd);
                }

                if (item.tag) {
                    const tag = document.createElement('span');
                    tag.className = 'action-end-tag';
                    tag.textContent = item.tag;
                    meta.appendChild(tag);
                }

                row.appendChild(mainCol);
                row.appendChild(meta);

                row.addEventListener('click', (e) => {
                    e.stopPropagation();
                    executeItem(item);
                });

                list.appendChild(row);
            });
        }

        function updateSelection() {
            const rows = list.querySelectorAll('.action-search-item');
            rows.forEach((r, idx) => {
                if (idx === activeIndex) {
                    r.classList.add('is-selected');
                    r.scrollIntoView({ block: 'nearest' });
                    input.setAttribute('aria-activedescendant', r.id);
                } else {
                    r.classList.remove('is-selected');
                }
            });
            if (activeIndex === -1) {
                input.removeAttribute('aria-activedescendant');
            }
        }

        function executeItem(item) {
            closeDropdown();
            if (typeof item.run === 'function') {
                item.run();
            }
        }

        function openDropdown() {
            dropdown.style.display = 'block';
            input.setAttribute('aria-expanded', 'true');
            computeAndRender();
        }

        function closeDropdown() {
            dropdown.style.display = 'none';
            input.setAttribute('aria-expanded', 'false');
            activeIndex = -1;
            updateSelection();
        }

        function computeAndRender() {
            const rawVal = input.value || '';
            const q = normalizeQuery(rawVal);

            if (rawVal.trim().length > 0) {
                if (iconSearch) iconSearch.classList.add('d-none');
                if (iconSend) iconSend.classList.remove('d-none');
                if (iconBox) iconBox.classList.add('is-active');
                if (kbdBadge) kbdBadge.style.opacity = '0';
                if (clearBtn) clearBtn.classList.remove('d-none');
            } else {
                if (iconSearch) iconSearch.classList.remove('d-none');
                if (iconSend) iconSend.classList.add('d-none');
                if (iconBox) iconBox.classList.remove('is-active');
                if (kbdBadge) kbdBadge.style.opacity = '1';
                if (clearBtn) clearBtn.classList.add('d-none');
            }

            if (!q) {
                renderItems(defaultActions, 'پیشنهادات و فیلترهای سریع لاگ');
                return;
            }

            const matchedActions = defaultActions.filter(act => {
                const normTitle = normalizeQuery(act.title);
                const normDesc = normalizeQuery(act.desc);
                return normTitle.includes(q) || normDesc.includes(q);
            });

            const matchedLogs = [];
            if (Array.isArray(allAuditLogs)) {
                for (let i = 0; i < allAuditLogs.length; i++) {
                    const l = allAuditLogs[i];
                    if (!l) continue;
                    const ip = normalizeQuery(l.ip_address || '');
                    const u = normalizeQuery(l.admin_username || '');
                    const title = normalizeQuery(l.admin_title || '');
                    const act = normalizeQuery(l.action || '');
                    const tid = l.target_report_id ? String(l.target_report_id) : '';
                    const track = normalizeQuery(l.target_tracking_code || '');

                    if (ip.includes(q) || u.includes(q) || title.includes(q) || act.includes(q) || tid.includes(q) || track.includes(q)) {
                        const actName = (ADMIN_FILTER_DESCRIPTIONS['audit-filter-action'] && ADMIN_FILTER_DESCRIPTIONS['audit-filter-action'][l.action]) ? ADMIN_FILTER_DESCRIPTIONS['audit-filter-action'][l.action] : l.action;
                        const iconColor = l.action === 'LOGIN' ? 'action-icon-blue' : (l.action.includes('REVOKE') || l.action.includes('DELETE') ? 'action-icon-rose' : 'action-icon-amber');
                        matchedLogs.push({
                            id: `log-${l.id || i}`,
                            title: `${l.admin_username || 'سیستم'} · ${actName}`,
                            desc: `IP: ${l.ip_address || '-'} ${l.target_tracking_code ? `· رهگیری: ${toPersianDigits(l.target_tracking_code)}` : (l.target_report_id ? `· پرونده #${toPersianDigits(l.target_report_id)}` : '')}`,
                            tag: l.admin_title || 'رویداد',
                            short: 'انتخاب',
                            iconBoxClass: iconColor,
                            iconSvg: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>',
                            run: () => {
                                input.value = l.target_tracking_code || l.ip_address || l.admin_username || '';
                                if (clearBtn) clearBtn.classList.remove('d-none');
                                applyAuditFiltersAndPagination(true);
                            }
                        });
                        if (matchedLogs.length >= 6) break;
                    }
                }
            }

            const combined = [...matchedActions, ...matchedLogs];
            renderItems(combined, `یافته‌ها در لاگ‌ها برای «${rawVal.trim()}»`);
        }

        input.addEventListener('focus', () => {
            if (!wrapper.classList.contains('is-open')) openSearch();
            openDropdown();
        });

        input.addEventListener('click', () => {
            if (!wrapper.classList.contains('is-open')) openSearch();
            if (dropdown.style.display === 'none') openDropdown();
        });

        input.addEventListener('input', () => {
            computeAndRender();
        });

        input.addEventListener('keydown', (e) => {
            if (dropdown.style.display !== 'none' && currentItems.length > 0) {
                if (e.key === 'ArrowDown') {
                    e.preventDefault();
                    activeIndex = (activeIndex + 1) % currentItems.length;
                    updateSelection();
                    return;
                }
                if (e.key === 'ArrowUp') {
                    e.preventDefault();
                    activeIndex = (activeIndex - 1 + currentItems.length) % currentItems.length;
                    updateSelection();
                    return;
                }
                if (e.key === 'Enter') {
                    if (activeIndex >= 0 && activeIndex < currentItems.length) {
                        e.preventDefault();
                        executeItem(currentItems[activeIndex]);
                        return;
                    }
                    closeDropdown();
                    return;
                }
            }
            if (e.key === 'Escape') {
                e.preventDefault();
                closeDropdown();
                if (input.value.trim() !== '') {
                    input.value = '';
                    if (typeof applyAuditFiltersAndPagination === 'function') applyAuditFiltersAndPagination(true);
                }
                closeSearch(false);
            }
        });

        document.addEventListener('click', (e) => {
            if (!wrapper.contains(e.target)) {
                closeDropdown();
                if (wrapper.classList.contains('is-open') && input.value.trim() === '') {
                    closeSearch(false);
                }
            }
        });

        if (input.value.trim().length > 0) {
            wrapper.classList.add('is-open');
            if (iconSearchToggle) iconSearchToggle.classList.add('d-none');
            if (iconCloseToggle) iconCloseToggle.classList.remove('d-none');
            if (clearBtn) clearBtn.classList.remove('d-none');
            if (kbdBadge) kbdBadge.style.opacity = '0';
        }

        window.closeAuditSearch = closeSearch;
    }

    initAuditActionSearchBar();

    btnRefreshAudit?.addEventListener('click', async () => {
        const icon = btnRefreshAudit.querySelector('.refresh-icon');
        btnRefreshAudit.disabled = true;
        if (icon) icon.classList.add('spinning');
        try {
            await fetchAuditLogs();
        } finally {
            setTimeout(() => {
                if (icon) icon.classList.remove('spinning');
                btnRefreshAudit.disabled = false;
            }, 300);
        }
    });

    auditFilterSearch?.addEventListener('input', debounce(() => {
        requestAnimationFrame(() => applyAuditFiltersAndPagination(true));
    }, 300));
    auditFilterAction?.addEventListener('change', () => applyAuditFiltersAndPagination(true));
    auditFilterUser?.addEventListener('change', () => applyAuditFiltersAndPagination(true));

    auditClearFiltersBtn?.addEventListener('click', () => {
        if (typeof window.closeAuditSearch === 'function') {
            window.closeAuditSearch(true);
        } else {
            if (auditFilterSearch) auditFilterSearch.value = '';
        }
        if (auditFilterAction) auditFilterAction.value = 'all';
        if (auditFilterUser) auditFilterUser.value = 'all';
        if (typeof window.resetAuditShamsiDateRangePicker === 'function') {
            window.resetAuditShamsiDateRangePicker();
        } else {
            auditActiveShamsiDateRange = null;
        }
        syncAllAdminCustomSelects();
        applyAuditFiltersAndPagination(true);
    });

    document.getElementById('btn-audit-first')?.addEventListener('click', () => { auditCurrentPage = 1; applyAuditFiltersAndPagination(false); });
    document.getElementById('btn-audit-prev')?.addEventListener('click', () => { if (auditCurrentPage > 1) { auditCurrentPage--; applyAuditFiltersAndPagination(false); } });
    document.getElementById('btn-audit-next')?.addEventListener('click', () => { auditCurrentPage++; applyAuditFiltersAndPagination(false); });
    document.getElementById('btn-audit-last')?.addEventListener('click', () => {
        const totalPages = Math.ceil(filteredAuditLogs.length / auditPageSize) || 1;
        auditCurrentPage = totalPages;
        applyAuditFiltersAndPagination(false);
    });

    const auditTbody = document.getElementById('adv-audit-tbody');
    if (auditTbody && !auditTbody._delegated) {
        auditTbody._delegated = true;
        auditTbody.addEventListener('click', async (e) => {
            const copyBtn = e.target.closest('[data-copy-tracking]');
            if (copyBtn) {
                const code = copyBtn.getAttribute('data-copy-tracking');
                if (code) {
                    try {
                        if (navigator.clipboard && navigator.clipboard.writeText) {
                            await navigator.clipboard.writeText(code);
                        } else {
                            const temp = document.createElement('textarea');
                            temp.value = code;
                            document.body.appendChild(temp);
                            temp.select();
                            document.execCommand('copy');
                            document.body.removeChild(temp);
                        }
                        showToast('کد رهگیری کپی شد');
                    } catch (err) {
                        showToast('خطا در کپی');
                    }
                }
            }
        });
    }

    function renderAuditTableSkeleton(count = 6) {
        const tbody = document.getElementById('adv-audit-tbody');
        if (!tbody) return;
        let html = '';
        for (let i = 0; i < count; i++) {
            html += `
                <tr class="skeleton-tr" aria-hidden="true">
                    <td><div class="skeleton-shimmer" style="width:24px; height:18px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:85px; height:18px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:95px; height:20px; border-radius:10px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:180px; height:16px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:90px; height:16px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:65px; height:18px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:110px; height:16px; border-radius:4px;"></div></td>
                </tr>
            `;
        }
        tbody.innerHTML = html;
    }

    async function fetchAuditLogs() {
        const tbody = document.getElementById('adv-audit-tbody');
        if (!tbody) return;
        renderAuditTableSkeleton(6);

        try {
            const res = await adminFetch('/api/admin/audit-logs');
            if (res.status === 403) {
                tbody.innerHTML = '<tr><td colspan="7" style="text-align:center; padding:20px; color:var(--color-danger);">دسترسی به لاگ‌های بازرسی محدود به پشتیبانی سامانه است.</td></tr>';
                return;
            }
            if (!res.ok) throw new Error('خطا در دریافت لاگ‌ها');

            allAuditLogs = await res.json();
            populateAuditUsersDropdown(allAuditLogs);
            applyAuditFiltersAndPagination(true);
        } catch (e) {

            tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:20px; color:var(--color-danger);">خطا: ${escapeHTML(e.message)}</td></tr>`;
        }
    }

    function renderAuditTable(logs) {
        const tbody = document.getElementById('adv-audit-tbody');
        if (!tbody) return;
        tbody.innerHTML = '';

        if (!logs || logs.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" style="text-align:center; padding:20px; color:var(--text-muted);">هیچ لاگ بازرسی با فیلترهای انتخابی یافت نشد.</td></tr>';
            return;
        }

        const fragment = document.createDocumentFragment();

        logs.forEach(l => {
            const tr = document.createElement('tr');
            const trackingCell = l.target_tracking_code ? `
                <div class="audit-tracking-cell">
                    <code class="audit-tracking-code ltr">${escapeHTML(l.target_tracking_code)}</code>
                    <button type="button" class="btn-mini-copy" data-copy-tracking="${escapeHTML(l.target_tracking_code)}" data-tooltip="کپی کد رهگیری" aria-label="کپی کد رهگیری">
                        <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                        </svg>
                    </button>
                </div>
            ` : (l.target_report_id ? `<span class="ltr" style="font-family:Tahoma,sans-serif; color:var(--text-muted); font-size:12px;">#${toPersianNum(l.target_report_id)}</span>` : '<span style="color:var(--text-subtle);">-</span>');

            tr.innerHTML = `
                <td>#${toPersianNum(l.id)}</td>
                <td class="ltr" style="text-align:right;">${formatJalaliDateTime(l.created_at, true)}</td>
                <td><strong class="ltr" style="font-family:Tahoma,sans-serif; color:var(--text-main);">${escapeHTML(l.admin_username)}</strong></td>
                <td>${escapeHTML(l.admin_title)}</td>
                <td><span class="badge-action-audit">${escapeHTML(l.action)}</span></td>
                <td>${trackingCell}</td>
                <td class="ltr" style="font-family:monospace; color:var(--text-muted);">${escapeHTML(l.ip_address)}</td>
            `;
            fragment.appendChild(tr);
        });

        tbody.appendChild(fragment);
        tbody._delegated = true;
    }

    const btnRefreshSys = document.getElementById('btn-refresh-system-status');
    btnRefreshSys?.addEventListener('click', () => fetchSystemStatus());

    async function fetchSystemStatus() {
        try {
            const res = await adminFetch('/api/admin/system-status');
            if (res.status === 403) return;
            if (!res.ok) throw new Error('خطا در دریافت وضعیت سامانه');

            const data = await res.json();

            const dbStatusEl = document.getElementById('sys-db-status');
            const dbConnsEl = document.getElementById('sys-db-conns');
            const dbIdleEl = document.getElementById('sys-db-idle');
            const dbInUseEl = document.getElementById('sys-db-inuse');
            const dbTzEl = document.getElementById('sys-db-tz');
            const dbTimeEl = document.getElementById('sys-db-time');
            const activeSessEl = document.getElementById('sys-active-sessions');
            const totalRepEl = document.getElementById('sys-total-reports');
            const unreadRepEl = document.getElementById('sys-unread-reports');
            const totalUsersEl = document.getElementById('sys-total-users');
            const totalLogsEl = document.getElementById('sys-total-logs');

            if (dbStatusEl) dbStatusEl.textContent = data.db_status;
            if (dbConnsEl) dbConnsEl.textContent = toPersianNum(data.db_open_conns);
            if (dbIdleEl) dbIdleEl.textContent = toPersianNum(data.db_idle_conns);
            if (dbInUseEl) dbInUseEl.textContent = toPersianNum(data.db_in_use_conns);
            if (dbTzEl) dbTzEl.textContent = data.db_timezone || 'Asia/Tehran';
            if (dbTimeEl) dbTimeEl.textContent = formatJalaliDateTime(data.db_time, true);
            if (activeSessEl) activeSessEl.textContent = toPersianNum(data.active_sessions);
            if (totalRepEl) totalRepEl.textContent = toPersianNum(data.total_reports);
            if (unreadRepEl) unreadRepEl.textContent = toPersianNum(data.unread_reports);
            if (totalUsersEl) totalUsersEl.textContent = toPersianNum(data.total_users);
            if (totalLogsEl) totalLogsEl.textContent = toPersianNum(data.total_audit_logs);

            showToast('وضعیت سلامت سامانه بروزرسانی شد');
        } catch (e) {
            showToast('خطا در دریافت تله‌متری سیستم');
        }
    }

    const btnRefreshSessions = document.getElementById('btn-refresh-active-sessions');
    btnRefreshSessions?.addEventListener('click', () => fetchActiveSessions());

    let allActiveSessions = [];

    function renderSessionsTableSkeleton(count = 5) {
        const tbody = document.getElementById('adv-sessions-tbody');
        if (!tbody) return;
        let html = '';
        for (let i = 0; i < count; i++) {
            html += `
                <tr class="skeleton-tr" aria-hidden="true">
                    <td><div class="skeleton-shimmer" style="width:24px; height:18px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:80px; height:18px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:110px; height:18px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:75px; height:20px; border-radius:10px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:85px; height:16px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:90px; height:16px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:105px; height:16px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:105px; height:16px; border-radius:4px;"></div></td>
                    <td><div class="skeleton-shimmer" style="width:70px; height:26px; border-radius:6px;"></div></td>
                </tr>
            `;
        }
        tbody.innerHTML = html;
    }

    async function fetchActiveSessions() {
        const tbody = document.getElementById('adv-sessions-tbody');
        if (!tbody) return;
        renderSessionsTableSkeleton(5);

        try {
            const res = await adminFetch('/api/admin/sessions');
            if (res.status === 403) {
                tbody.innerHTML = '<tr><td colspan="9" style="text-align:center; padding:20px; color:var(--color-danger);">دسترسی به اطلاعات نشست‌ها محدود به پشتیبانی سامانه است.</td></tr>';
                return;
            }
            if (!res.ok) throw new Error('خطا در دریافت نشست‌های فعال');

            allActiveSessions = await res.json();
            renderSessionsTable(allActiveSessions);

            const totalEl = document.getElementById('adv-sess-total');
            const superEl = document.getElementById('adv-sess-super');
            const ipsEl = document.getElementById('adv-sess-ips');

            if (totalEl) totalEl.textContent = toPersianNum(allActiveSessions.length);
            if (superEl) superEl.textContent = toPersianNum(allActiveSessions.filter(s => s.is_superuser).length);
            const uniqueIPs = new Set(allActiveSessions.map(s => s.ip)).size;
            if (ipsEl) ipsEl.textContent = toPersianNum(uniqueIPs);

            showToast('لیست نشست‌های فعال بروزرسانی شد');
        } catch (e) {

            tbody.innerHTML = `<tr><td colspan="9" style="text-align:center; padding:20px; color:var(--color-danger);">خطا: ${escapeHTML(e.message)}</td></tr>`;
        }
    }

    function renderSessionsTable(sessions) {
        const tbody = document.getElementById('adv-sessions-tbody');
        if (!tbody) return;
        tbody.innerHTML = '';

        if (!sessions || sessions.length === 0) {
            tbody.innerHTML = '<tr><td colspan="9" style="text-align:center; padding:20px; color:var(--text-muted);">هیچ نشست فعالی در حافظه سرور یافت نشد.</td></tr>';
            return;
        }

        const fragment = document.createDocumentFragment();

        sessions.forEach(s => {
            const tr = document.createElement('tr');
            const isSelf = currentAdmin && currentAdmin.username === s.username;

            const ttlMinutes = Math.floor(s.ttl_seconds / 60);
            const ttlSecs = s.ttl_seconds % 60;
            const ttlDisplay = `${toPersianNum(ttlMinutes)} دقیقه و ${toPersianNum(ttlSecs)} ثانیه`;

            const uaSummary = parseUserAgent(s.user_agent);

            tr.innerHTML = `
                <td>
                    <span class="badge" style="font-family:monospace; background:rgba(15,23,42,0.8); border:1px solid var(--border-subtle); color:#38bdf8;">
                        ${escapeHTML(s.session_id)}
                    </span>
                    ${isSelf ? '<div style="margin-top:4px;"><span class="badge badge-read">نشست فعلی شما</span></div>' : ''}
                </td>
                <td>
                    <strong style="color:var(--text-main); font-weight:600;">${escapeHTML(s.full_name || s.username)}</strong>
                    <div style="font-size:12px; color:var(--text-muted); margin-top:2px;">${escapeHTML(s.role_title || '-')}</div>
                </td>
                <td>
                    <span class="ltr" style="font-family:Tahoma,sans-serif; color:var(--text-muted);">${escapeHTML(s.username)}</span>
                    <div style="margin-top:3px;">
                        ${s.is_superuser ? '<span class="badge-super">پشتیبانی سامانه</span>' : '<span class="badge-standard">کارشناس</span>'}
                    </div>
                </td>
                <td>
                    <span class="ltr" style="font-family:monospace; color:#86efac; font-weight:600;">${escapeHTML(s.ip)}</span>
                </td>
                <td style="max-width: 240px;">
                    <div style="font-size:12.5px; color:#e2e8f0; font-weight:500;">${escapeHTML(uaSummary.client)}</div>
                    <div class="ltr" style="font-size:11px; color:var(--text-muted); white-space:nowrap; overflow:hidden; text-overflow:ellipsis; max-width:230px;" data-tooltip="${escapeHTML(s.user_agent)}">
                        ${escapeHTML(s.user_agent)}
                    </div>
                </td>
                <td class="ltr" style="text-align:right; font-size:12px;">
                    ${formatJalaliDateTime(s.created_at, true)}
                </td>
                <td class="ltr" style="text-align:right; font-size:12px;">
                    ${formatJalaliDateTime(s.last_active, true)}
                </td>
                <td>
                    <span class="badge" style="background:rgba(245,158,11,0.15); color:#fbbf24; border:1px solid rgba(245,158,11,0.3); font-weight:600; display:inline-flex; align-items:center; gap:4px;">
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                        ${ttlDisplay}
                    </span>
                </td>
                <td>
                    ${isSelf ? `
                        <span style="font-size:12px; color:var(--text-muted);">غیرقابل لغو</span>
                    ` : `
                        <button type="button" class="btn-action btn-danger" data-action="terminate-session" data-token="${escapeHTML(s.token)}" data-user="${escapeHTML(s.username)}" style="font-size:12px; padding:5px 10px;">
                            خاتمه نشست
                        </button>
                    `}
                </td>
            `;

            fragment.appendChild(tr);
        });

        tbody.appendChild(fragment);

        if (!tbody._delegated) {
            tbody._delegated = true;
            tbody.addEventListener('click', (e) => {
                const b = e.target.closest('button[data-action="terminate-session"]');
                if (!b) return;
                const token = b.getAttribute('data-token');
                const username = b.getAttribute('data-user');
                terminateSession(token, username);
            });
        }
    }

    async function terminateSession(token, username) {
        const confirmed = await showConfirmDialog({
            title: 'قطع دسترسی و خاتمه فوری نشست',
            description: `آیا از قطع دسترسی و خاتمه فوری نشست کاربر «${username}» اطمینان دارید؟ نشست کاربر بلافاصله باطل شده و از سامانه خارج می‌گردد.`,
            confirmText: 'بله، خاتمه نشست',
            cancelText: 'انصراف',
            variant: 'danger',
            iconType: 'shield-alert'
        });
        if (!confirmed) {
            return;
        }

        try {
            const res = await adminFetch('/api/admin/sessions/terminate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ token: token })
            });

            if (res.ok) {
                showToast(`نشست کاربر "${username}" با موفقیت قطع و لغو گردید`);
                fetchActiveSessions();
            } else {
                const err = await res.text();
                showToast(`خطا: ${err}`);
            }
        } catch (e) {
            showToast('خطای ارتباط با سرور.');
        }
    }

    function parseUserAgent(ua) {
        if (!ua) return { client: 'نامشخص' };
        let browser = 'مرورگر ناشناس';
        let os = 'سیستم‌عامل ناشناس';

        if (ua.includes('Windows NT 10.0')) os = 'ویندوز ۱۰ / ۱۱';
        else if (ua.includes('Windows')) os = 'ویندوز';
        else if (ua.includes('Linux')) os = 'لینوکس';
        else if (ua.includes('Macintosh') || ua.includes('Mac OS')) os = 'macOS';
        else if (ua.includes('Android')) os = 'اندروید';
        else if (ua.includes('iPhone') || ua.includes('iPad')) os = 'iOS';

        if (ua.includes('Edg/')) browser = 'Microsoft Edge';
        else if (ua.includes('Chrome/') && !ua.includes('Edg/')) browser = 'Google Chrome';
        else if (ua.includes('Firefox/')) browser = 'Mozilla Firefox';
        else if (ua.includes('Safari/') && !ua.includes('Chrome/')) browser = 'Apple Safari';

        return { client: `${browser} روی ${os}` };
    }

    let allNotifications = [];
    let previousNotifUnreadCount = -1;
    const btnNotifBell = document.getElementById('btnNotifBell');
    const notifDropdown = document.getElementById('notifDropdown');
    const notifBadgeCount = document.getElementById('notifBadgeCount');
    const notifUnreadTag = document.getElementById('notifUnreadTag');
    const notifListContainer = document.getElementById('notifListContainer');
    const btnMarkAllNotifsRead = document.getElementById('btnMarkAllNotifsRead');

    async function fetchNotifications() {
        if (!dashboardSection || dashboardSection.style.display === 'none' || dashboardSection.classList.contains('d-none') || !currentAdmin) return;

        try {
            const res = await adminFetch('/api/admin/notifications');
            if (res.status === 401) {
                forceLogoutOnConnectionDrop();
                return;
            }
            if (!res.ok) return;

            const data = await res.json();
            allNotifications = data.notifications || [];
            const unreadCount = data.unread_count || 0;

            if (notifBadgeCount) {
                if (unreadCount > 0) {
                    notifBadgeCount.textContent = toPersianNum(unreadCount);
                    notifBadgeCount.classList.remove('d-none');
                } else {
                    notifBadgeCount.textContent = '۰';
                    notifBadgeCount.classList.add('d-none');
                }
            }

            if (notifUnreadTag) {
                notifUnreadTag.textContent = `${toPersianNum(unreadCount)} جدید`;
            }

            if (previousNotifUnreadCount !== -1 && unreadCount > previousNotifUnreadCount) {
                if (btnNotifBell) {
                    btnNotifBell.classList.remove('bell-shake');
                    void btnNotifBell.offsetWidth;
                    btnNotifBell.classList.add('bell-shake');
                }
                fetchReports(true, false).catch(() => { });
            }
            previousNotifUnreadCount = unreadCount;

            if (notifDropdown && notifDropdown.style.display === 'block') {
                renderNotificationsList(allNotifications);
            }
        } catch (e) {
            console.error('Failed to fetch notifications', e);
        }
    }

    function renderNotificationsList(notifs) {
        if (!notifListContainer) return;
        notifListContainer.innerHTML = '';

        if (!notifs || notifs.length === 0) {
            notifListContainer.innerHTML = `
                <div class="notif-empty-state">
                    <div class="notif-empty-icon">
                        <svg viewBox="0 0 24 24" width="38" height="38" fill="none" stroke="currentColor" stroke-width="1.5">
                            <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
                            <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
                        </svg>
                    </div>
                    <div class="notif-empty-text">هیچ اعلانی در سیستم ثبت نشده است.</div>
                </div>
            `;
            return;
        }

        if (notifListContainer && !notifListContainer._perfDelegated) {
            notifListContainer._perfDelegated = true;
            notifListContainer.addEventListener('click', async (e) => {
                const item = e.target.closest('.notif-item');
                if (!item) return;

                const notifId = parseInt(item.getAttribute('data-id'), 10);
                const targetReportId = item.getAttribute('data-report-id');
                const isUnread = item.classList.contains('unread');

                if (isUnread && !isNaN(notifId)) {
                    try {
                        await adminFetch('/api/admin/notifications/mark-read', {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({ id: notifId })
                        });
                        item.classList.remove('unread');
                        item.querySelector('.notif-dot-unread')?.remove();
                        const nObj = allNotifications.find(n => n.id === notifId);
                        if (nObj) nObj.is_read = true;
                        fetchNotifications();
                    } catch (err) { }
                }

                if (notifDropdown) {
                    notifDropdown.style.display = 'none';
                    notifDropdown.classList.remove('is-open');
                }

                if (targetReportId) {
                    const repIdNum = parseInt(targetReportId, 10);
                    if (!isNaN(repIdNum)) {
                        switchMainTab('reports');
                        viewReport(repIdNum);
                    }
                }
            });
        }

        const fragment = document.createDocumentFragment();

        notifs.forEach(n => {
            const item = document.createElement('div');
            item.className = `notif-item ${!n.is_read ? 'unread' : ''}`;
            item.setAttribute('data-id', n.id);
            if (n.target_report_id) {
                item.setAttribute('data-report-id', n.target_report_id);
            }

            const isReply = n.notif_type === 'new_reply';
            const isForward = n.notif_type === 'report_forwarded';
            const isRegionalResponse = n.notif_type === 'regional_response';
            let iconHTML = '';
            if (isRegionalResponse) {
                iconHTML = `
                    <div class="notif-icon-box notif-icon-regional" style="background: rgba(99, 102, 241, 0.15); color: #818cf8;">
                        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"></path>
                            <circle cx="9" cy="7" r="4"></circle>
                            <polyline points="16 11 18 13 22 9"></polyline>
                        </svg>
                    </div>
                `;
            } else if (isForward) {
                iconHTML = `
                    <div class="notif-icon-box notif-icon-forward" style="background: rgba(14, 165, 233, 0.15); color: #38bdf8;">
                        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                            <path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"></path>
                            <polyline points="16 6 12 2 8 6"></polyline>
                            <line x1="12" y1="2" x2="12" y2="15"></line>
                        </svg>
                    </div>
                `;
            } else if (isReply) {
                iconHTML = `
                    <div class="notif-icon-box notif-icon-reply">
                        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                        </svg>
                    </div>
                `;
            } else {
                iconHTML = `
                    <div class="notif-icon-box notif-icon-report">
                        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
                            <polyline points="14 2 14 8 20 8"></polyline>
                        </svg>
                    </div>
                `;
            }

            item.innerHTML = `
                ${iconHTML}
                <div class="notif-content">
                    <div class="notif-item-header">
                        <div class="notif-item-title">${escapeHTML(n.title)}</div>
                        ${!n.is_read ? '<span class="notif-dot-unread" data-tooltip="خوانده‌نشده"></span>' : ''}
                    </div>
                    <div class="notif-item-msg">${escapeHTML(n.message)}</div>
                    <div class="notif-item-meta">
                        <span>${formatJalaliDateTime(n.created_at, true)}</span>
                        ${n.target_tracking_code ? `<span class="ltr font-mono" style="color: #38bdf8;">${escapeHTML(n.target_tracking_code)}</span>` : ''}
                    </div>
                </div>
            `;

            fragment.appendChild(item);
        });

        notifListContainer.appendChild(fragment);
    }

    if (btnNotifBell) {
        btnNotifBell.addEventListener('click', (e) => {
            e.stopPropagation();
            if (!notifDropdown) return;
            const isShown = notifDropdown.style.display === 'block';
            if (isShown) {
                notifDropdown.style.display = 'none';
                notifDropdown.classList.remove('is-open');
            } else {
                notifDropdown.style.display = 'block';
                notifDropdown.classList.add('is-open');
                renderNotificationsList(allNotifications);
                fetchNotifications();
            }
        });
    }

    document.addEventListener('click', (e) => {
        if (notifDropdown && notifDropdown.style.display === 'block') {
            if (!notifDropdown.contains(e.target) && !btnNotifBell.contains(e.target)) {
                notifDropdown.style.display = 'none';
                notifDropdown.classList.remove('is-open');
            }
        }
    });

    const btnUserMenu = document.getElementById('btnUserMenu');
    const userMenuDropdown = document.getElementById('userMenuDropdown');

    if (btnUserMenu && userMenuDropdown) {
        btnUserMenu.addEventListener('click', (e) => {
            e.stopPropagation();
            const isOpen = userMenuDropdown.style.display === 'block';
            if (isOpen) {
                userMenuDropdown.style.display = 'none';
                btnUserMenu.classList.remove('is-active');
            } else {
                userMenuDropdown.style.display = 'block';
                btnUserMenu.classList.add('is-active');
                if (notifDropdown) {
                    notifDropdown.style.display = 'none';
                    notifDropdown.classList.remove('is-open');
                }
            }
        });

        document.addEventListener('click', (e) => {
            if (userMenuDropdown.style.display === 'block') {
                if (!userMenuDropdown.contains(e.target) && !btnUserMenu.contains(e.target)) {
                    userMenuDropdown.style.display = 'none';
                    btnUserMenu.classList.remove('is-active');
                }
            }
        });

        userMenuDropdown.querySelectorAll('.user-menu-item').forEach(item => {
            item.addEventListener('click', () => {
                userMenuDropdown.style.display = 'none';
                btnUserMenu.classList.remove('is-active');
            });
        });
    }

    if (btnMarkAllNotifsRead) {
        btnMarkAllNotifsRead.addEventListener('click', async (e) => {
            e.stopPropagation();
            try {
                const res = await adminFetch('/api/admin/notifications/mark-all-read', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' }
                });
                if (res.ok) {
                    showToast('تمامی اعلان‌ها خوانده شدند');
                    await fetchNotifications();
                }
            } catch (e) {
                showToast('خطا در علامت‌گذاری اعلان‌ها');
            }
        });
    }

    let cachedForwardingRecipients = null;
    let selectedRecipientUsernames = new Set();
    let currentForwardReportId = null;

    const modalForward = document.getElementById('modal-forward-report');
    const btnCloseForward = document.getElementById('btn-close-forward-modal');
    const btnCancelForward = document.getElementById('btn-cancel-forward');
    const formForward = document.getElementById('form-forward-report');
    const forwardSearchInput = document.getElementById('forward-search-input');
    const btnSelectAllRecipients = document.getElementById('btn-select-all-recipients');
    const btnDeselectAllRecipients = document.getElementById('btn-deselect-all-recipients');
    const forwardSelectedCount = document.getElementById('forward-selected-count');
    const forwardRecipientsContainer = document.getElementById('forward-recipients-container');
    const forwardStatusMsg = document.getElementById('forward-status-msg');
    const forwardModalTrackingCode = document.getElementById('forward-modal-tracking-code');

    function closeForwardModal() {
        if (modalForward) modalForward.style.display = 'none';
        selectedRecipientUsernames.clear();
        currentForwardReportId = null;
        if (forwardSearchInput) forwardSearchInput.value = '';
        if (forwardStatusMsg) forwardStatusMsg.style.display = 'none';
    }

    btnCloseForward?.addEventListener('click', closeForwardModal);
    btnCancelForward?.addEventListener('click', closeForwardModal);
    modalForward?.addEventListener('click', (e) => {
        if (e.target === modalForward) closeForwardModal();
    });

    async function fetchForwardingRecipients() {
        try {
            const res = await adminFetch('/api/admin/forwarding-recipients');
            if (res.ok) {
                cachedForwardingRecipients = await res.json();
                return cachedForwardingRecipients;
            }
        } catch (e) {
            console.error('Failed to fetch forwarding recipients:', e);
        }
        return [];
    }

    function updateSelectedRecipientsCounter() {
        if (forwardSelectedCount) {
            forwardSelectedCount.textContent = toPersianNum(selectedRecipientUsernames.size);
        }
    }

    function renderRecipientList(searchTerm = '') {
        if (!forwardRecipientsContainer) return;
        forwardRecipientsContainer.innerHTML = '';

        if (!cachedForwardingRecipients || cachedForwardingRecipients.length === 0) {
            forwardRecipientsContainer.innerHTML = `
                <div style="text-align:center; padding:25px; color:var(--text-muted); font-size:13px;">
                    هیچ کاربری با نقش «نواحی و مناطق» در سامانه تعریف نشده است.
                </div>
            `;
            return;
        }

        const normSearch = normalizePersianForSearch(searchTerm.trim());
        const filtered = cachedForwardingRecipients.filter(u => {
            if (!normSearch) return true;
            const normName = normalizePersianForSearch(u.full_name || '');
            const normTitle = normalizePersianForSearch(u.role_title || '');
            return normName.includes(normSearch) || normTitle.includes(normSearch);
        });

        if (filtered.length === 0) {
            forwardRecipientsContainer.innerHTML = `
                <div style="text-align:center; padding:20px; color:var(--text-muted); font-size:13px;">
                    کارشناسی با مشخصات جستجو شده یافت نشد.
                </div>
            `;
            return;
        }

        const fragment = document.createDocumentFragment();

        filtered.forEach(u => {
            const item = document.createElement('label');
            const isChecked = selectedRecipientUsernames.has(u.username);
            item.className = `forward-recipient-item ${isChecked ? 'selected' : ''}`;
            item.style.display = 'flex';
            item.style.alignItems = 'center';
            item.style.justifyContent = 'space-between';
            item.style.padding = '10px 14px';
            item.style.margin = '4px 0';
            item.style.borderRadius = 'var(--radius-md, 8px)';
            item.style.border = '1px solid var(--border-subtle, rgba(255,255,255,0.08))';
            item.style.background = isChecked ? 'rgba(56, 189, 248, 0.12)' : 'rgba(255,255,255,0.02)';
            item.style.cursor = 'pointer';
            item.style.transition = 'all 0.15s ease';

            item.innerHTML = `
                <div style="display: flex; align-items: center; gap: 12px;">
                    <input type="checkbox" class="recipient-checkbox" value="${escapeHTML(u.username)}" ${isChecked ? 'checked' : ''} style="width: 17px; height: 17px; accent-color: var(--color-primary, #0ea5e9); cursor: pointer;">
                    <div>
                        <div class="recipient-fullname" style="font-weight: 600; font-size: 13.5px; color: var(--text-main);">
                            ${escapeHTML(u.full_name || 'کارشناس حراست')}
                        </div>
                        <div class="recipient-roletitle" style="font-size: 11.5px; color: var(--text-muted); margin-top: 2px;">
                            ${escapeHTML(u.role_title || 'ناحیه / منطقه')}
                        </div>
                    </div>
                </div>
                <span class="badge-regional-pill" style="font-size: 11px; padding: 2px 8px; border-radius: 10px; background: rgba(14, 165, 233, 0.15); color: #38bdf8; border: 1px solid rgba(56, 189, 248, 0.3);">
                    نواحی و مناطق
                </span>
            `;

            const cb = item.querySelector('.recipient-checkbox');
            cb.addEventListener('change', (e) => {
                if (cb.checked) {
                    selectedRecipientUsernames.add(u.username);
                    item.classList.add('selected');
                    item.style.background = 'rgba(56, 189, 248, 0.12)';
                } else {
                    selectedRecipientUsernames.delete(u.username);
                    item.classList.remove('selected');
                    item.style.background = 'rgba(255,255,255,0.02)';
                }
                updateSelectedRecipientsCounter();
            });

            fragment.appendChild(item);
        });

        forwardRecipientsContainer.appendChild(fragment);
    }

    forwardSearchInput?.addEventListener('input', () => {
        renderRecipientList(forwardSearchInput.value);
    });

    btnSelectAllRecipients?.addEventListener('click', () => {
        if (!cachedForwardingRecipients) return;
        const normSearch = normalizePersianForSearch(forwardSearchInput ? forwardSearchInput.value.trim() : '');
        cachedForwardingRecipients.forEach(u => {
            if (!normSearch) {
                selectedRecipientUsernames.add(u.username);
            } else {
                const normName = normalizePersianForSearch(u.full_name || '');
                const normTitle = normalizePersianForSearch(u.role_title || '');
                if (normName.includes(normSearch) || normTitle.includes(normSearch)) {
                    selectedRecipientUsernames.add(u.username);
                }
            }
        });
        renderRecipientList(forwardSearchInput ? forwardSearchInput.value : '');
        updateSelectedRecipientsCounter();
    });

    btnDeselectAllRecipients?.addEventListener('click', () => {
        selectedRecipientUsernames.clear();
        renderRecipientList(forwardSearchInput ? forwardSearchInput.value : '');
        updateSelectedRecipientsCounter();
    });

    function renderRecipientListSkeleton(count = 4) {
        if (!forwardRecipientsContainer) return;
        let html = '';
        for (let i = 0; i < count; i++) {
            html += `
                <div class="recipient-select-item" style="display:flex; justify-content:space-between; align-items:center; padding:10px 12px; border-bottom:1px solid rgba(255,255,255,0.05); pointer-events:none;">
                    <div style="display:flex; align-items:center; gap:10px;">
                        <div class="skeleton-shimmer" style="width:16px; height:16px; border-radius:4px;"></div>
                        <div style="display:flex; flex-direction:column; gap:6px;">
                            <div class="skeleton-shimmer" style="width:120px; height:14px; border-radius:4px;"></div>
                            <div class="skeleton-shimmer" style="width:80px; height:11px; border-radius:4px;"></div>
                        </div>
                    </div>
                    <div class="skeleton-shimmer" style="width:75px; height:20px; border-radius:10px;"></div>
                </div>
            `;
        }
        forwardRecipientsContainer.innerHTML = html;
    }

    async function openForwardModal(reportId, trackingCode) {
        currentForwardReportId = reportId;
        selectedRecipientUsernames.clear();
        updateSelectedRecipientsCounter();

        if (forwardModalTrackingCode) {
            forwardModalTrackingCode.textContent = trackingCode || `#${reportId}`;
        }
        if (forwardStatusMsg) forwardStatusMsg.style.display = 'none';
        if (forwardSearchInput) forwardSearchInput.value = '';

        if (modalForward) {

            const card = modalForward.querySelector('.modal-card') || modalForward;
            promoteTemporarily(card, 'transform, opacity', 250);
            modalForward.style.display = 'flex';
        }

        if (!cachedForwardingRecipients) {
            renderRecipientListSkeleton(4);
        }

        await fetchForwardingRecipients();
        renderRecipientList();
    }

    formForward?.addEventListener('submit', async (e) => {
        e.preventDefault();
        if (!currentForwardReportId) return;

        if (selectedRecipientUsernames.size === 0) {
            if (forwardStatusMsg) {
                forwardStatusMsg.textContent = 'لطفاً حداقل یک کارشناس از نواحی و مناطق را برای ارجاع خبر انتخاب فرمایید.';
                forwardStatusMsg.className = 'response-box text-danger';
                forwardStatusMsg.style.display = 'block';
            } else {
                showToast('لطفاً حداقل یک کارشناس را انتخاب فرمایید', 3000, 'error');
            }
            return;
        }

        const submitBtn = document.getElementById('btn-submit-forward');
        const btnText = submitBtn?.querySelector('.btn-text');
        const btnSpinner = submitBtn?.querySelector('.btn-spinner');

        if (submitBtn) submitBtn.disabled = true;
        if (btnText) btnText.textContent = 'در حال ثبت ارجاع...';
        if (btnSpinner) btnSpinner.classList.remove('d-none');
        if (forwardStatusMsg) forwardStatusMsg.style.display = 'none';

        try {
            const res = await adminFetch('/api/admin/report/forward', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    report_id: currentForwardReportId,
                    recipient_usernames: Array.from(selectedRecipientUsernames)
                })
            });

            if (res.ok) {
                const count = selectedRecipientUsernames.size;
                showToast(`خبر با موفقیت به ${toPersianNum(count)} کارشناس ارجاع داده شد`, 3000, 'success');
                closeForwardModal();
            } else {
                const err = await res.text();
                if (forwardStatusMsg) {
                    forwardStatusMsg.textContent = err || 'خطا در ارجاع خبر.';
                    forwardStatusMsg.className = 'response-box text-danger';
                    forwardStatusMsg.style.display = 'block';
                } else {
                    showToast(err || 'خطا در ارجاع خبر', 3500, 'error');
                }
            }
        } catch (err) {
            if (forwardStatusMsg) {
                forwardStatusMsg.textContent = 'خطای ارتباط با سرور.';
                forwardStatusMsg.className = 'response-box text-danger';
                forwardStatusMsg.style.display = 'block';
            } else {
                showToast('خطای ارتباط با سرور', 3500, 'error');
            }
        } finally {
            if (submitBtn) submitBtn.disabled = false;
            if (btnText) btnText.textContent = 'تایید و ارجاع خبر';
            if (btnSpinner) btnSpinner.classList.add('d-none');
        }
    });

    const modalHistory = document.getElementById('modal-forward-history');
    const btnCloseHistory = document.getElementById('btn-close-history-modal');
    const btnCloseHistoryFooter = document.getElementById('btn-close-history-footer');
    const historyModalTrackingCode = document.getElementById('history-modal-tracking-code');
    const forwardHistoryBody = document.getElementById('forward-history-body');

    function closeHistoryModal() {
        if (modalHistory) modalHistory.style.display = 'none';
        if (forwardHistoryBody) forwardHistoryBody.innerHTML = '';
    }

    btnCloseHistory?.addEventListener('click', closeHistoryModal);
    btnCloseHistoryFooter?.addEventListener('click', closeHistoryModal);
    modalHistory?.addEventListener('click', (e) => {
        if (e.target === modalHistory) closeHistoryModal();
    });

    function renderForwardHistorySkeleton() {
        if (!forwardHistoryBody) return;
        forwardHistoryBody.innerHTML = `
            <div class="skeleton-flowchart-container" aria-hidden="true">
                <div class="skeleton-flowchart-node">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <div class="skeleton-shimmer" style="width: 120px; height: 18px; border-radius: 6px;"></div>
                        <div class="skeleton-shimmer" style="width: 70px; height: 20px; border-radius: 10px;"></div>
                    </div>
                    <div class="skeleton-shimmer" style="width: 80%; height: 14px; border-radius: 4px;"></div>
                    <div class="skeleton-shimmer" style="width: 50%; height: 13px; border-radius: 4px;"></div>
                </div>
                <div class="skeleton-flowchart-connector"></div>
                <div class="skeleton-flowchart-node">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <div class="skeleton-shimmer" style="width: 130px; height: 18px; border-radius: 6px;"></div>
                        <div class="skeleton-shimmer" style="width: 75px; height: 20px; border-radius: 10px;"></div>
                    </div>
                    <div class="skeleton-shimmer" style="width: 85%; height: 14px; border-radius: 4px;"></div>
                    <div class="skeleton-shimmer" style="width: 60%; height: 13px; border-radius: 4px;"></div>
                </div>
            </div>
        `;
    }

    async function openForwardHistoryModal(reportId, trackingCode) {
        if (historyModalTrackingCode) {
            historyModalTrackingCode.textContent = trackingCode || `#${reportId}`;
        }
        if (modalHistory) modalHistory.style.display = 'flex';

        renderForwardHistorySkeleton();

        try {
            const res = await adminFetch(`/api/admin/report/forward-history?id=${reportId}`);
            if (!res.ok) {
                const err = await res.text();
                if (forwardHistoryBody) {
                    forwardHistoryBody.innerHTML = `
                        <div class="flow-empty-state" style="text-align:center; padding:35px; color:var(--color-danger, #ef4444);">
                            خطا در دریافت تاریخچه: ${escapeHTML(err)}
                        </div>
                    `;
                }
                return;
            }

            const historyList = await res.json();
            renderForwardHistoryFlowchart(historyList);
        } catch (e) {
            console.error('Failed to load or render forward history:', e);
            if (forwardHistoryBody) {
                forwardHistoryBody.innerHTML = `
                    <div class="flow-empty-state" style="text-align:center; padding:35px; color:var(--color-danger, #ef4444);">
                        خطا در پردازش تاریخچه ارجاع: ${escapeHTML(e.message || 'خطای ناشناخته')}
                    </div>
                `;
            }
        }
    }

    function renderForwardHistoryFlowchart(items) {
        if (!forwardHistoryBody) return;
        forwardHistoryBody.innerHTML = '';

        if (!items || items.length === 0) {
            forwardHistoryBody.innerHTML = `
                <div class="flow-empty-state" style="text-align:center; padding:45px 20px; color:var(--text-muted);">
                    <div style="width:54px; height:54px; border-radius:50%; background:rgba(255,255,255,0.04); display:flex; align-items:center; justify-content:center; margin:0 auto 12px; color:var(--text-muted);">
                        <svg viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="1.6">
                            <circle cx="12" cy="12" r="10"></circle>
                            <polyline points="12 6 12 12 16 14"></polyline>
                        </svg>
                    </div>
                    <div style="font-weight:600; font-size:14.5px; color:var(--text-main); margin-bottom:6px;">تاکنون ارجاعی برای این پرونده ثبت نگردیده است</div>
                    <div style="font-size:12.5px;">گزارش در کارتابل اداره کل بررسی و در صورت نیاز به کارشناسان نواحی و مناطق ارجاع خواهد شد.</div>
                </div>
            `;
            return;
        }

        const groups = [];
        items.forEach(item => {
            const fwdAt = item.forwarded_at || '';
            const key = `${item.forwarded_by}_${fwdAt.length >= 16 ? fwdAt.substring(0, 16) : fwdAt}`;
            let group = groups.find(g => g.key === key);
            if (!group) {

                let resolvedForwarderName = item.forwarder_name || item.forwarded_by_name;
                if (!resolvedForwarderName || resolvedForwarderName === item.forwarded_by) {
                    const matchUser = (typeof allAdminUsers !== 'undefined' && Array.isArray(allAdminUsers)) ? allAdminUsers.find(u => u.username === item.forwarded_by) : null;
                    if (matchUser && matchUser.full_name) {
                        resolvedForwarderName = matchUser.full_name;
                    } else if (item.forwarder_title || item.forwarded_by_title) {
                        resolvedForwarderName = item.forwarder_title || item.forwarded_by_title;
                    } else {
                        resolvedForwarderName = 'کارشناس اداره کل';
                    }
                }

                group = {
                    key: key,
                    forwarder_name: resolvedForwarderName,
                    forwarder_title: item.forwarder_title || item.forwarded_by_title || 'اداره کل حراست',
                    forwarded_at: item.forwarded_at,
                    recipients: []
                };
                groups.push(group);
            }
            group.recipients.push(item);
        });

        const container = document.createElement('div');
        container.className = 'flowchart-wrapper';

        groups.forEach((group, idx) => {
            const groupCard = document.createElement('div');
            groupCard.className = 'flow-group-card';
            if (idx > 0) {
                groupCard.style.marginTop = '28px';
                groupCard.style.borderTop = '1px dashed var(--border-subtle, rgba(255,255,255,0.12))';
                groupCard.style.paddingTop = '24px';
            }

            const forwardDateFa = formatJalaliDateTime(group.forwarded_at, true);

            const sourceHtml = `
                <div class="flow-node-source-container">
                    <div class="flow-node flow-node-source">
                        <div class="flow-node-header">
                            <div class="flow-node-badge-role">مقام ارجاع‌دهنده (اداره کل)</div>
                        </div>
                        <div class="flow-node-body">
                            <div class="flow-node-icon-box source-icon">
                                <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                                    <circle cx="12" cy="7" r="4"></circle>
                                </svg>
                            </div>
                            <div class="flow-node-details">
                                <div class="flow-person-name">${escapeHTML(group.forwarder_name)}</div>
                                <div class="flow-person-title">${escapeHTML(group.forwarder_title)}</div>
                            </div>
                        </div>
                        <div class="flow-node-footer">
                            <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2">
                                <circle cx="12" cy="12" r="10"></circle>
                                <polyline points="12 6 12 12 16 14"></polyline>
                            </svg>
                            <span>تاریخ و ساعت ارجاع: <strong class="font-tahoma ltr">${forwardDateFa}</strong></span>
                        </div>
                    </div>
                </div>
            `;

            const connectorHtml = `
                <div class="flow-connector-wrapper">
                    <div class="flow-connector-line"></div>
                    <div class="flow-connector-arrow">
                        <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                            <line x1="12" y1="5" x2="12" y2="19"></line>
                            <polyline points="19 12 12 19 5 12"></polyline>
                        </svg>
                    </div>
                    <div class="flow-connector-label">ارجاع به ${toPersianNum(group.recipients.length)} کارشناس</div>
                </div>
            `;

            let destinationsHtml = `<div class="flow-destinations-grid count-${Math.min(group.recipients.length, 4)}">`;

            group.recipients.forEach(r => {
                const isViewed = r.is_viewed;
                const viewDateFa = isViewed && r.viewed_at ? formatJalaliDateTime(r.viewed_at, true) : null;

                const seenBadge = isViewed ? `
                    <div class="flow-seen-badge seen">
                        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                            <polyline points="20 6 9 17 4 12"></polyline>
                        </svg>
                        <div class="seen-text-wrap">
                            <span class="seen-title">مشاهده شد</span>
                            <span class="seen-time font-tahoma ltr">${viewDateFa}</span>
                        </div>
                    </div>
                ` : `
                    <div class="flow-seen-badge unseen">
                        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <circle cx="12" cy="12" r="10"></circle>
                            <line x1="12" y1="8" x2="12" y2="12"></line>
                            <line x1="12" y1="16" x2="12.01" y2="16"></line>
                        </svg>
                        <div class="seen-text-wrap">
                            <span class="seen-title">مشاهده نشده</span>
                            <span class="seen-sub">در انتظار رویت کارشناس</span>
                        </div>
                    </div>
                `;

                const isMainAdmin = currentAdmin && (currentAdmin.role === 'edare_kol' || currentAdmin.role === 'superuser' || currentAdmin.is_superuser);
                const revokeBtnHtml = isMainAdmin ? `
                    <!-- [HARDEN-11] Accessible revoke forward button with aria-hidden icon -->
                    <button type="button" class="btn-revoke-forward" data-fwd-id="${r.id}" data-rep-id="${r.report_id}" data-recip-name="${escapeHTML(r.recipient_name || r.recipient_username)}" data-recip-user="${escapeHTML(r.recipient_username)}" title="حذف ارجاع و قطع دسترسی این کارشناس" aria-label="حذف ارجاع">
                        <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <line x1="18" y1="6" x2="6" y2="18"></line>
                            <line x1="6" y1="6" x2="18" y2="18"></line>
                        </svg>
                    </button>
                ` : '';

                const responsePreviewHtml = r.response_text ? `
                    <div class="flow-regional-response-preview">
                        <div class="response-preview-header">
                            <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
                            <span>گزارش بررسی کارشناس (${r.responded_at ? formatJalaliDateTime(r.responded_at, true) : 'ثبت‌شده'}):</span>
                        </div>
                        <div class="response-preview-body">${escapeHTML(r.response_text)}</div>
                    </div>
                ` : '';

                destinationsHtml += `
                    <div class="flow-node flow-node-destination ${isViewed ? 'is-viewed' : 'is-unseen'}">
                        <div class="flow-node-header">
                            <div class="flow-node-badge-role recipient">دریافت‌کننده (ناحیه / منطقه)</div>
                            ${revokeBtnHtml}
                        </div>
                        <div class="flow-node-body">
                            <div class="flow-node-icon-box dest-icon ${isViewed ? 'viewed' : 'unseen'}">
                                <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
                                    <circle cx="9" cy="7" r="4"></circle>
                                    <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
                                    <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
                                </svg>
                            </div>
                            <div class="flow-node-details">
                                <div class="flow-person-name">${escapeHTML(r.recipient_name || 'کارشناس حراست')}</div>
                                <div class="flow-person-title">${escapeHTML(r.recipient_title || 'حراست ناحیه / منطقه')}</div>
                            </div>
                        </div>
                        ${responsePreviewHtml}
                        <div class="flow-node-footer status-footer">
                            ${seenBadge}
                        </div>
                    </div>
                `;
            });

            destinationsHtml += `</div>`;

            groupCard.innerHTML = sourceHtml + connectorHtml + destinationsHtml;
            container.appendChild(groupCard);
        });

        forwardHistoryBody.appendChild(container);
        if (typeof convertTitles === 'function') {
            convertTitles(forwardHistoryBody);
        } else if (typeof window !== 'undefined' && typeof window.convertTitles === 'function') {
            window.convertTitles(forwardHistoryBody);
        }

        if (!forwardHistoryBody._delegated) {
            forwardHistoryBody._delegated = true;
            forwardHistoryBody.addEventListener('click', async (e) => {
                const btn = e.target.closest('.btn-revoke-forward');
                if (!btn) return;
                e.stopPropagation();
                const fwdId = btn.getAttribute('data-fwd-id');
                const repId = btn.getAttribute('data-rep-id');
                const recipName = btn.getAttribute('data-recip-name') || 'این کارشناس';
                const recipUser = btn.getAttribute('data-recip-user');

                const confirmed = await showConfirmDialog({
                    title: 'حذف ارجاع پرونده',
                    description: `آیا از حذف ارجاع این پرونده به «${recipName}» اطمینان دارید؟ با این اقدام، دسترسی این کارشناس به این پرونده فوراً قطع شده و کلیه اعلان‌های مرتبط از کارتابل وی پاک خواهند شد.`,
                    confirmText: 'بله، حذف ارجاع',
                    cancelText: 'انصراف',
                    variant: 'warning',
                    iconType: 'shield-alert'
                });
                if (!confirmed) {
                    return;
                }

                btn.disabled = true;
                btn.style.opacity = '0.4';

                try {
                    const res = await adminFetch('/api/admin/report/remove-forward', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({
                            forwarding_id: parseInt(fwdId, 10),
                            report_id: parseInt(repId, 10),
                            recipient_username: recipUser
                        })
                    });

                    if (!res.ok) {
                        const err = await res.text();
                        throw new Error(err || 'خطا در حذف ارجاع');
                    }

                    const resData = await res.json();
                    showToast(resData.message || 'ارجاع پرونده با موفقیت لغو گردید.');

                    openForwardHistoryModal(parseInt(repId, 10), historyModalTrackingCode?.textContent || '');

                    if (currentViewingReportId && currentViewingReportId === parseInt(repId, 10)) {
                        viewReport(currentViewingReportId, true);
                    }
                } catch (err) {
                    showToast(err.message || 'خطای ارتباط با سرور در حذف ارجاع', 4000, 'error');
                    btn.disabled = false;
                    btn.style.opacity = '';
                }
            });
        }
    }

    document.addEventListener('visibilitychange', () => {
        if (document.hidden) {
            clearInterval(timerInterval);
            stopBackgroundPolling();
            if (_sseConnection) { _sseConnection.close(); _sseConnection = null; }
        } else {
            if (dashboardSection && dashboardSection.style.display !== 'none') {
                if (typeof resumeSessionTimer === 'function') resumeSessionTimer();
                connectAdminSSE();
                if (typeof fetchNotifications === 'function') fetchNotifications();
            }
        }
    });

    runWhenIdle(() => {
        if ('IntersectionObserver' in window) {
            const animObserver = new IntersectionObserver((entries) => {
                entries.forEach(e => {
                    e.target.style.animationPlayState = e.isIntersecting ? 'running' : 'paused';
                });
            });
            document.querySelectorAll('.beam, .corner-glow, .kpi-pulse-dot, .badge-dot-beacon, .pulsing-halo').forEach(el => animObserver.observe(el));
        }
    });

    document.addEventListener('pointerdown', (e) => {
        const btn = e.target.closest('.btn-send-animated');
        if (!btn || btn.disabled) return;
        if (e.pointerType === 'touch') {
            btn.classList.add('is-animating');
            if (btn._touchAnimTimeout) clearTimeout(btn._touchAnimTimeout);
            btn._touchAnimTimeout = setTimeout(() => {
                btn.classList.remove('is-animating');
                btn._touchAnimTimeout = null;
            }, 1000);
        }
    }, { passive: true });
});
