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
        .finally(() => clearTimeout(timer));
}

(function initPreloader() {
    const PRELOADER_MIN_MS = 3000;
    const startTime = performance.now();
    let pageLoaded = false;
    let dismissed = false;

    function attemptDismiss() {
        if (dismissed || !pageLoaded) return;
        const elapsed = performance.now() - startTime;
        if (elapsed < PRELOADER_MIN_MS) {
            setTimeout(attemptDismiss, PRELOADER_MIN_MS - elapsed);
            return;
        }
        dismissed = true;
        const preloader = document.getElementById('appPreloader') || document.getElementById('adminPreloader');
        const skeleton = document.getElementById('pageSkeleton') || document.getElementById('adminSkeleton');
        if (preloader) {
            preloader.classList.add('fade-out');
            setTimeout(() => preloader.remove(), 450);
        }
        if (skeleton) skeleton.remove();
    }

    function onComplete() {
        pageLoaded = true;
        attemptDismiss();
    }

    if (document.readyState === 'complete') {
        onComplete();
    } else {
        window.addEventListener('load', onComplete);
        document.addEventListener('DOMContentLoaded', onComplete);
    }

    setTimeout(() => {
        onComplete();
    }, 10000);
})();

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

    function isMobileOrTouch() {
        if (window.innerWidth <= 768) return true;
        try {
            if (window.matchMedia && (window.matchMedia('(hover: none)').matches || window.matchMedia('(pointer: coarse)').matches)) {
                return true;
            }
        } catch (e) {}
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

        const side = rect.top < 46 ? 'bottom' : 'top';
        tip.setAttribute('data-side', side);

        tip.style.left = '-9999px';
        tip.style.top = '-9999px';
        tip.style.visibility = 'hidden';
        tip.style.display = 'block';

        const tipWidth = tip.offsetWidth;
        const tipHeight = tip.offsetHeight;

        let left = rect.left + (rect.width / 2) - (tipWidth / 2);
        left = Math.max(10, Math.min(window.innerWidth - tipWidth - 10, left));

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

try { initAppTooltips(); } catch (e) { console.error('Tooltips init failed:', e); }

document.addEventListener('DOMContentLoaded', () => {

    const toPersianNum = (n) => String(n).replace(/[0-9]/g, d => '۰۱۲۳۴۵۶۷۸۹'[d]);

    const formatFileSize = (bytes) => {
        if (bytes === 0) return '۰ بایت';
        const k = 1024;
        const sizes = ['بایت', 'کیلوبایت', 'مگابایت'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        const formatted = parseFloat((bytes / Math.pow(k, i)).toFixed(1));
        return `${toPersianNum(formatted)} ${sizes[i]}`;
    };

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

    const cachedJalaliDateOnly = new Intl.DateTimeFormat('fa-IR', {
        timeZone: 'Asia/Tehran',
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
    });

    const cachedJalaliTimeOnly = new Intl.DateTimeFormat('fa-IR', {
        timeZone: 'Asia/Tehran',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
    });

    const formatJalaliDateTime = (isoStr, includeSeconds = false) => {
        if (!isoStr) return 'نامشخص';
        try {
            const dateObj = parseTimestamp(isoStr);
            if (!dateObj) return 'نامشخص';
            return (includeSeconds ? cachedJalaliFormatterFull : cachedJalaliFormatterShort).format(dateObj);
        } catch (e) {
            return 'نامشخص';
        }
    };

    function populateOfficialReceipt(info) {
        if (!info) return;
        const code = info.tracking_code || '';
        const now = info.timestamp ? (parseTimestamp(info.timestamp) || new Date()) : new Date();

        const dateEl = document.getElementById('receipt-date');
        const timeEl = document.getElementById('receipt-time');
        const trackCodeEl = document.getElementById('receipt-tracking-code');
        const categoryEl = document.getElementById('receipt-category');
        const typeEl = document.getElementById('receipt-type');
        const reporterEl = document.getElementById('receipt-reporter');
        const phoneEl = document.getElementById('receipt-phone');
        const attachmentsEl = document.getElementById('receipt-attachments');
        const barcodeNumEl = document.getElementById('receipt-barcode-num');

        if (trackCodeEl) trackCodeEl.textContent = code;
        if (barcodeNumEl) barcodeNumEl.textContent = code;
        if (categoryEl) categoryEl.textContent = info.category || 'عمومی و سایر موضوعات';
        if (typeEl) typeEl.textContent = info.type || 'محرمانه';
        if (reporterEl) reporterEl.textContent = info.reporter || 'محفوظ (ثبت ناشناس)';
        if (phoneEl) phoneEl.textContent = info.phone || 'ثبت نشده';
        if (attachmentsEl) attachmentsEl.textContent = info.attachments || 'فاقد فایل پیوست';

        if (dateEl) {
            try {
                dateEl.textContent = cachedJalaliDateOnly.format(now);
            } catch (e) {
                dateEl.textContent = 'امروز';
            }
        }

        if (timeEl) {
            try {
                timeEl.textContent = cachedJalaliTimeOnly.format(now);
            } catch (e) {
                timeEl.textContent = '--:--:--';
            }
        }
    }

    const toastContainer = document.getElementById('toast-container');
    const TOAST_ICONS = {
        success: '<svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><path d="m9 11 3 3L22 4"/></svg>',
        error: '<svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/></svg>',
        warning: '<svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>',
        info: '<svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>'
    };

    function showToast(message, type = 'info', duration = 4000) {
        if (!toastContainer) return;
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;

        toast.innerHTML = `
            <span class="toast-icon">${TOAST_ICONS[type] || TOAST_ICONS.info}</span>
            <span class="toast-body"></span>
            <button class="toast-close" aria-label="بستن">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
            </button>
        `;
        toast.querySelector('.toast-body').textContent = message;
        toast.querySelector('.toast-close').addEventListener('click', () => dismissToast(toast));
        toastContainer.appendChild(toast);
        if (duration > 0) {
            setTimeout(() => dismissToast(toast), duration);
        }
    }

    function dismissToast(toast) {
        if (!toast || !toast.parentNode) return;
        toast.classList.add('toast-exit');
        setTimeout(() => toast.remove(), 300);
    }

    let offlineBanner = null;
    function showOfflineBanner() {
        if (offlineBanner) return;
        offlineBanner = document.createElement('div');
        offlineBanner.className = 'offline-banner';
        offlineBanner.innerHTML = `
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m2 2 20 20"/><path d="M8.5 16.5a5 5 0 0 1 7 0"/><path d="M2 8.82a15 15 0 0 1 4.17-2.65"/><path d="M10.66 5c4.01-.36 8.14.9 11.34 3.76"/><path d="M16.85 11.25a10 10 0 0 1 2.22 1.68"/><path d="M5 13a10 10 0 0 1 5.24-2.78"/><line x1="12" y1="20" x2="12.01" y2="20"/></svg>
            <span>ارتباط با سرور قطع شده است. لطفاً اتصال اینترنت خود را بررسی کنید.</span>
        `;
        document.body.prepend(offlineBanner);
    }

    function hideOfflineBanner() {
        if (!offlineBanner) return;
        offlineBanner.className = 'offline-banner online-restored';
        offlineBanner.innerHTML = `
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h.01"/><path d="M2 8.82a15 15 0 0 1 20 0"/><path d="M5 12.86a10 10 0 0 1 14 0"/><path d="M8.5 16.5a5 5 0 0 1 7 0"/></svg>
            <span>ارتباط با سرور برقرار شد.</span>
        `;
        setTimeout(() => { if (offlineBanner) { offlineBanner.remove(); offlineBanner = null; } }, 3000);
    }

    window.addEventListener('offline', showOfflineBanner);
    window.addEventListener('online', hideOfflineBanner);
    if (!navigator.onLine) showOfflineBanner();

    const DRAFT_KEY = 'herasat_report_draft';
    const MIN_REPORT_DESC_LENGTH = 50;
    const draftIndicator = document.getElementById('draft-saved-indicator');
    let draftSaveTimer = null;

    function saveDraft() {
        try {
            const desc = document.getElementById('description')?.value?.trim() || '';

            if (!desc) {
                clearDraft();
                return;
            }

            const draft = {
                reporter_name: document.getElementById('reporter_name')?.value || '',
                reporter_phone: document.getElementById('reporter_phone')?.value || '',
                description: document.getElementById('description')?.value || ''
            };

            sessionStorage.setItem(DRAFT_KEY, JSON.stringify(draft));

            if (draftIndicator) {
                draftIndicator.classList.add('visible');
                setTimeout(() => draftIndicator.classList.remove('visible'), 1500);
            }
        } catch (e) {  }
    }

    function restoreDraft() {
        try {
            const saved = sessionStorage.getItem(DRAFT_KEY);
            if (!saved) return;
            const draft = JSON.parse(saved);

            if (!draft.description || !draft.description.trim()) {
                clearDraft();
                return;
            }

            if (draft.reporter_name) { const el = document.getElementById('reporter_name'); if (el) el.value = draft.reporter_name; }
            if (draft.reporter_phone) { const el = document.getElementById('reporter_phone'); if (el) el.value = draft.reporter_phone; }

            const catEl = document.getElementById('category');
            if (catEl) {
                catEl.value = '';
                catEl.dispatchEvent(new Event('sync-custom'));
            }

            if (draft.description) {
                const el = document.getElementById('description');
                if (el) {
                    el.value = draft.description;
                    el.dispatchEvent(new Event('input', { bubbles: true }));
                }
            }
            showToast('پیش‌نویس قبلی شما بازیابی شد.', 'info', 3000);
        } catch (e) {  }
    }

    function clearDraft() {
        try { sessionStorage.removeItem(DRAFT_KEY); } catch (e) {  }
    }

    function scheduleDraftSave() {
        clearTimeout(draftSaveTimer);
        draftSaveTimer = setTimeout(saveDraft, 1500);
    }

    const form = document.getElementById('reportForm');
    const fileInput = document.getElementById('evidence');
    const dropzone = document.getElementById('dropzone');
    const dropzonePrompt = document.getElementById('dropzone-prompt');
    const filePreviewCard = document.getElementById('file-preview-card');
    const filePreviewName = document.getElementById('file-preview-name');
    const filePreviewSize = document.getElementById('file-preview-size');
    const btnRemoveFile = document.getElementById('btn-remove-file');

    const submitBtn = document.getElementById('submitBtn');
    const responseBox = document.getElementById('responseMessage');
    const descTextarea = document.getElementById('description');
    const charCounter = document.getElementById('char-counter');
    const currentCharCount = document.getElementById('current-char-count');

    let currentStep = 1;
    const totalSteps = 3;
    const steps = document.querySelectorAll('.form-step');
    const nextBtns = document.querySelectorAll('.btn-next');
    const prevBtns = document.querySelectorAll('.btn-prev');
    const stepTrackers = document.querySelectorAll('.step-item');

    function makeTextareaAutoExpand(el) {
        if (!el) return;
        el.style.resize = 'none';
        el.style.overflowY = 'hidden';
        const minHeight = el.id === 'description' ? 130 : 60;
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

    function updateStepTracker() {
        steps.forEach((step, index) => {
            const stepNum = index + 1;
            if (stepNum === currentStep) {
                step.classList.add('active');

                step.querySelectorAll('textarea').forEach(ta => {
                    ta.style.height = 'auto';
                    ta.style.height = Math.max(ta.scrollHeight, 130) + 'px';
                });
            } else {
                step.classList.remove('active');
            }
        });

        for (let i = 1; i <= totalSteps; i++) {
            const tracker = document.getElementById(`tracker-${i}`);
            const connector = document.getElementById(`connector-${i}`);
            const circle = tracker?.querySelector('.step-circle');

            if (i < currentStep) {
                tracker?.classList.add('completed');
                tracker?.classList.remove('active');
                tracker?.removeAttribute('aria-current');
                tracker?.setAttribute('tabindex', '0');
                tracker?.setAttribute('role', 'button');
                tracker?.setAttribute('aria-label', `بازگشت به مرحله ${toPersianNum(i)}`);
                if (circle) circle.innerHTML = '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="20 6 9 17 4 12"></polyline></svg>';
                if (connector) connector.classList.add('completed');
            } else if (i === currentStep) {
                tracker?.classList.add('active');
                tracker?.classList.remove('completed');
                tracker?.setAttribute('aria-current', 'step');
                tracker?.removeAttribute('tabindex');
                tracker?.removeAttribute('role');
                tracker?.removeAttribute('aria-label');
                if (circle) circle.innerHTML = `<span class="step-num">${toPersianNum(i)}</span>`;
                if (connector) connector.classList.remove('completed');
            } else {
                tracker?.classList.remove('active', 'completed');
                tracker?.removeAttribute('aria-current');
                tracker?.removeAttribute('tabindex');
                tracker?.removeAttribute('role');
                tracker?.removeAttribute('aria-label');
                if (circle) circle.innerHTML = `<span class="step-num">${toPersianNum(i)}</span>`;
                if (connector) connector.classList.remove('completed');
            }
        }

        window.scrollTo({ top: 0, behavior: getScrollBehavior() });

        if (currentStep === 3) {
            const submitCaptchaImg = document.getElementById('submit-captcha-img');
            if (!submitCaptchaImg || !submitCaptchaImg.src || !submitCaptchaImg.src.startsWith('data:')) {
                loadCaptcha('submit-captcha-img', 'submit_captcha_id', 'submit_captcha_answer');
            }
        }

        const activeStep = document.getElementById(`step-${currentStep}`);
        if (activeStep) {
            const firstInput = activeStep.querySelector('input:not([type=hidden]):not([type=file]), select, textarea');
            if (firstInput) setTimeout(() => firstInput.focus(), 100);
        }
    }

    function validateCurrentStep() {
        const activeStepEl = document.getElementById(`step-${currentStep}`);
        if (!activeStepEl) return true;

        if (currentStep === 2) {
            const categorySelect = document.getElementById('category');
            if (!categorySelect || !categorySelect.value) {
                const trigger = document.getElementById('categoryDropdownTrigger');
                if (trigger) {
                    trigger.focus();
                    trigger.classList.add('error-pulse');
                    setTimeout(() => trigger.classList.remove('error-pulse'), 1200);
                }
                showToast('لطفاً موضوع گزارش را از منوی موضوعات انتخاب فرمایید.', 'warning');
                return false;
            }

            const descInput = document.getElementById('description');
            const descVal = descInput?.value?.trim() || '';
            if (!descVal) {
                if (descInput) {
                    descInput.focus();
                    descInput.classList.add('error-pulse');
                    setTimeout(() => descInput.classList.remove('error-pulse'), 1200);
                }
                showToast('لطفاً شرح کامل گزارش را وارد فرمایید.', 'warning');
                return false;
            }

            if (descVal.length < MIN_REPORT_DESC_LENGTH) {
                if (descInput) {
                    descInput.focus();
                    descInput.classList.add('error-pulse');
                    setTimeout(() => descInput.classList.remove('error-pulse'), 1200);
                }
                showToast(`متن شرح گزارش بسیار کوتاه است؛ لطفاً حداقل ۵۰ کاراکتر شرح و جزئیات وارد فرمایید (تعداد فعلی: ${toPersianNum(descVal.length)} کاراکتر).`, 'warning');
                return false;
            }
        }

        const inputs = activeStepEl.querySelectorAll('input:not(.native-select-hidden), select:not(.native-select-hidden), textarea');

        for (let input of inputs) {
            if (!input.checkValidity()) {
                input.focus();
                input.classList.add('error-pulse');
                setTimeout(() => input.classList.remove('error-pulse'), 1200);

                let msg = 'لطفاً فیلد مورد نظر را به درستی تکمیل فرمایید.';
                if (input.id === 'description') {
                    msg = 'لطفاً شرح کامل گزارش را وارد فرمایید.';
                } else if (input.id === 'captcha_answer') {
                    msg = 'لطفاً کد امنیتی را وارد فرمایید.';
                } else {
                    const label = activeStepEl.querySelector(`label[for="${input.id}"]`);
                    if (label) {
                        const labelText = label.textContent.replace('*', '').replace('(اختیاری)', '').trim();
                        msg = `لطفاً ${labelText} را تکمیل فرمایید.`;
                    }
                }
                showToast(msg, 'warning');
                return false;
            }
        }

        return true;
    }

    function initCategoryDropdown() {
        const dropdown = document.getElementById('categoryCustomDropdown');
        const trigger = document.getElementById('categoryDropdownTrigger');
        const menu = document.getElementById('categoryDropdownMenu');
        const nativeSelect = document.getElementById('category');
        const label = document.getElementById('categorySelectedLabel');
        const iconContainer = document.getElementById('categorySelectedIcon');
        const items = menu?.querySelectorAll('.dropdown-item');

        if (!dropdown || !trigger || !menu || !nativeSelect) return;

        function toggleMenu(open) {
            const isOpen = open !== undefined ? open : !dropdown.classList.contains('open');
            if (isOpen) {
                dropdown.classList.add('open');
                trigger.setAttribute('aria-expanded', 'true');
            } else {
                dropdown.classList.remove('open');
                trigger.setAttribute('aria-expanded', 'false');
            }
        }

        trigger.addEventListener('click', (e) => {
            e.stopPropagation();
            toggleMenu();
        });

        items.forEach(item => {
            item.addEventListener('click', (e) => {
                e.stopPropagation();
                const val = item.getAttribute('data-value');
                const itemLabel = item.querySelector('.item-label')?.textContent;
                const iconSvg = item.querySelector('.item-icon-box svg')?.cloneNode(true);

                nativeSelect.value = val;
                nativeSelect.dispatchEvent(new Event('change', { bubbles: true }));
                nativeSelect.dispatchEvent(new Event('input', { bubbles: true }));

                if (label) label.textContent = itemLabel || val;
                if (iconContainer && iconSvg) {
                    iconContainer.innerHTML = '';
                    iconContainer.appendChild(iconSvg);
                }

                items.forEach(i => i.classList.remove('selected'));
                item.classList.add('selected');

                toggleMenu(false);
            });
        });

        function syncFromNative() {
            const currentVal = nativeSelect.value;
            if (!currentVal) {
                if (label) label.textContent = '-- موضوع گزارش را انتخاب کنید --';
                if (iconContainer) {
                    iconContainer.innerHTML = '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 6h16M4 12h16M4 18h7"></path></svg>';
                }
                items.forEach(i => i.classList.remove('selected'));
                return;
            }
            const matchingItem = Array.from(items).find(i => i.getAttribute('data-value') === currentVal);
            if (matchingItem) {
                const itemLabel = matchingItem.querySelector('.item-label')?.textContent;
                const iconSvg = matchingItem.querySelector('.item-icon-box svg')?.cloneNode(true);
                if (label) label.textContent = itemLabel || currentVal;
                if (iconContainer && iconSvg) {
                    iconContainer.innerHTML = '';
                    iconContainer.appendChild(iconSvg);
                }
                items.forEach(i => i.classList.remove('selected'));
                matchingItem.classList.add('selected');
            }
        }

        nativeSelect.addEventListener('sync-custom', syncFromNative);
        syncFromNative();

        document.addEventListener('click', (e) => {
            if (!dropdown.contains(e.target)) {
                toggleMenu(false);
            }
        });

        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && dropdown.classList.contains('open')) {
                toggleMenu(false);
                trigger.focus();
            }
        });
    }

    initCategoryDropdown();

    nextBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            if (validateCurrentStep() && currentStep < totalSteps) {
                currentStep++;
                updateStepTracker();
            }
        });
    });

    prevBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            if (currentStep > 1) {
                currentStep--;
                updateStepTracker();
            }
        });
    });

    stepTrackers.forEach(item => {
        const handleStepNav = () => {
            const targetStep = parseInt(item.getAttribute('data-step'), 10);
            if (targetStep && targetStep < currentStep) {
                currentStep = targetStep;
                updateStepTracker();
            }
        };
        item.addEventListener('click', handleStepNav);

        item.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                handleStepNav();
            }
        });
    });

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey && document.getElementById('submitSection')?.contains(document.activeElement)) {
            const activeEl = document.activeElement;
            if (activeEl && activeEl.tagName === 'TEXTAREA') return;
            if (activeEl && activeEl.type === 'submit') return;
            e.preventDefault();
            if (currentStep < totalSteps && validateCurrentStep()) {
                currentStep++;
                updateStepTracker();
            }
        }
    });

    ['reporter_name', 'reporter_phone', 'description'].forEach(id => {
        const el = document.getElementById(id);
        if (el) {
            el.addEventListener('input', scheduleDraftSave);
            el.addEventListener('change', scheduleDraftSave);
        }
    });

    restoreDraft();

    if (descTextarea && currentCharCount) {
        const updateCharCount = () => {
            const length = descTextarea.value.length;
            currentCharCount.textContent = toPersianNum(length);

            if (length > 4500) {
                charCounter?.classList.add('danger');
                charCounter?.classList.remove('warning');
            } else if (length > 4000) {
                charCounter?.classList.add('warning');
                charCounter?.classList.remove('danger');
            } else {
                charCounter?.classList.remove('warning', 'danger');
            }
        };

        descTextarea.addEventListener('input', updateCharCount);
        updateCharCount();
    }

    const MAX_FILE_SIZE = 20 * 1024 * 1024;
    const ALLOWED_EXTENSIONS = ['.pdf', '.jpg', '.jpeg', '.png', '.zip', '.rar'];

    function getBeFileIconSvg(fileName, mimeType = '') {
        const ext = fileName.includes('.') ? fileName.split('.').pop().toLowerCase() : '';
        if (['jpg', 'jpeg', 'png', 'webp', 'gif'].includes(ext) || mimeType.startsWith('image/')) {
            return `<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="18" x="3" y="3" rx="2" ry="2"/><circle cx="9" cy="9" r="2"/><path d="m21 15-3.086-3.086a2 2 0 0 0-2.828 0L6 21"/></svg>`;
        }
        if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext) || mimeType.includes('zip') || mimeType.includes('compressed')) {
            return `<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 12v-2"/><path d="M10 18v-2"/><path d="M10 6V4"/><path d="M14 2v20"/><path d="M15.5 22H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h11.5"/><path d="M14 6h2"/><path d="M14 10h2"/><path d="M14 14h2"/><path d="M14 18h2"/><path d="M4 10h2"/><path d="M4 14h2"/><path d="M4 18h2"/><path d="M4 6h2"/></svg>`;
        }
        if (ext === 'pdf' || mimeType.includes('pdf')) {
            return `<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/><path d="M10 9H8"/><path d="M16 13H8"/><path d="M16 17H8"/></svg>`;
        }
        return `<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/></svg>`;
    }

    const beFileQueue = document.getElementById('be-file-queue');
    const beFileSummary = document.getElementById('be-file-summary');
    const beFileSummaryCount = document.getElementById('be-file-summary-count');
    const beFileSummarySize = document.getElementById('be-file-summary-size');
    const beFileList = document.getElementById('be-file-list');

    let selectedFiles = [];

    const ESCAPE_MAP = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' };
    function escapeHtml(str) {
        if (str == null) return '';
        return String(str).replace(/[&<>"']/g, c => ESCAPE_MAP[c]);
    }

    function renderFileQueue(newlyAddedIndices = []) {
        if (!beFileQueue || !beFileList) return;

        if (selectedFiles.length === 0) {
            beFileQueue.classList.add('d-none');
            if (beFileSummary) beFileSummary.classList.add('d-none');
            beFileList.innerHTML = '';
            return;
        }

        beFileQueue.classList.remove('d-none');

        const totalBytes = selectedFiles.reduce((sum, f) => sum + f.size, 0);

        if (beFileSummary) {
            if (selectedFiles.length > 1) {
                beFileSummary.classList.remove('d-none');
                if (beFileSummaryCount) {
                    beFileSummaryCount.textContent = `${toPersianNum(selectedFiles.length)} فایل پیوست شده`;
                }
                if (beFileSummarySize) {
                    beFileSummarySize.textContent = `مجموع: ${formatFileSize(totalBytes)} (سقف: ۲۰ مگابایت)`;
                }
            } else {
                beFileSummary.classList.add('d-none');
            }
        }

        beFileList.innerHTML = '';

        selectedFiles.forEach((file, idx) => {
            const ext = file.name.includes('.') ? file.name.split('.').pop().toUpperCase() : 'FILE';
            const formattedMeta = `${formatFileSize(file.size)} · ${ext}`;

            const row = document.createElement('div');
            row.className = 'be-file-row';
            row.setAttribute('data-file-idx', idx);

            row.innerHTML = `
                <div class="be-file-main">
                    <div class="be-file-leading-icon">
                        ${getBeFileIconSvg(file.name, file.type)}
                    </div>
                    <div class="be-file-meta-group">
                        <p class="be-file-name" title="${escapeHtml(file.name)}" dir="auto">${escapeHtml(file.name)}</p>
                        <p class="be-file-meta">${formattedMeta}</p>
                    </div>
                    <div class="be-file-actions">
                        <span class="be-status-icon text-emerald" data-tooltip="فایل آماده ارسال است">
                            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor"
                                stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
                                <path d="m9 11 3 3L22 4"></path>
                            </svg>
                        </span>
                        <button type="button" class="be-btn-remove" aria-label="حذف فایل" data-tooltip="حذف این فایل" data-remove-idx="${idx}">
                            <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor"
                                stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                                <path d="M18 6 6 18"></path>
                                <path d="m6 6 12 12"></path>
                            </svg>
                        </button>
                    </div>
                </div>
                <div class="be-progress-bar" role="progressbar" aria-valuemin="0" aria-valuemax="100" aria-valuenow="100">
                    <div class="be-progress-fill" style="width: ${newlyAddedIndices.includes(idx) ? '0%' : '100%'};"></div>
                </div>
            `;

            beFileList.appendChild(row);
        });

        if (beFileList && !beFileList._perfDelegated) {
            beFileList._perfDelegated = true;
            beFileList.addEventListener('click', (e) => {
                const btn = e.target.closest('.be-btn-remove');
                if (!btn) return;
                e.preventDefault();
                e.stopPropagation();
                const idx = parseInt(btn.getAttribute('data-remove-idx'), 10);
                if (!isNaN(idx)) {
                    removeFile(idx);
                }
            });
        }

        if (newlyAddedIndices.length > 0) {
            requestAnimationFrame(() => {
                setTimeout(() => {
                    newlyAddedIndices.forEach(idx => {
                        const row = beFileList.querySelector(`.be-file-row[data-file-idx="${idx}"]`);
                        const fill = row?.querySelector('.be-progress-fill');
                        if (fill) fill.style.width = '100%';
                    });
                }, 30);
            });
        }
    }

    function addFiles(fileList) {
        if (!fileList || fileList.length === 0) return;

        const incomingFiles = Array.from(fileList);
        let newlyAddedIndices = [];
        let rejectedCount = 0;
        let duplicateCount = 0;
        let sizeExceeded = false;

        for (const file of incomingFiles) {
            const fileName = file.name.toLowerCase();
            const isValidExt = ALLOWED_EXTENSIONS.some(ext => fileName.endsWith(ext));

            if (!isValidExt) {
                rejectedCount++;
                continue;
            }

            const isDuplicate = selectedFiles.some(f => f.name === file.name && f.size === file.size && f.lastModified === file.lastModified);
            if (isDuplicate) {
                duplicateCount++;
                continue;
            }

            const currentTotal = selectedFiles.reduce((sum, f) => sum + f.size, 0);
            if (currentTotal + file.size > MAX_FILE_SIZE) {
                sizeExceeded = true;
                continue;
            }

            selectedFiles.push(file);
            newlyAddedIndices.push(selectedFiles.length - 1);
        }

        if (rejectedCount > 0) {
            showToast('برخی فایل‌ها به دلیل فرمت غیرمجاز رد شدند. فرمت‌های مجاز: PDF, JPG, PNG, ZIP, RAR', 'error', 4500);
        }

        if (sizeExceeded) {
            const currentTotal = selectedFiles.reduce((sum, f) => sum + f.size, 0);
            showToast(`افزودن برخی فایل‌ها به دلیل عبور مجموع حجم پیوست‌ها از سقف ۲۰ مگابایت انجام نشد (حجم فعلی: ${formatFileSize(currentTotal)}).`, 'warning', 5000);
        }

        if (duplicateCount > 0 && newlyAddedIndices.length === 0 && !sizeExceeded && rejectedCount === 0) {
            showToast('این فایل قبلاً به لیست پیوست‌ها افزوده شده است.', 'info', 3000);
        }

        renderFileQueue(newlyAddedIndices);
    }

    function removeFile(index) {
        const row = beFileList?.querySelector(`.be-file-row[data-file-idx="${index}"]`);
        if (row) {
            row.classList.add('removing');
            setTimeout(() => {
                selectedFiles.splice(index, 1);
                renderFileQueue();
                showToast('فایل پیوست حذف شد.', 'info', 2000);
            }, 180);
        } else {
            selectedFiles.splice(index, 1);
            renderFileQueue();
        }
    }

    function clearFileInput() {
        selectedFiles = [];
        if (fileInput) fileInput.value = '';
        renderFileQueue();
    }

    if (fileInput) {
        fileInput.addEventListener('change', (e) => {
            if (e.target.files && e.target.files.length > 0) {
                addFiles(e.target.files);
                fileInput.value = '';
            }
        });
    }

    if (dropzone) {
        dropzone.addEventListener('click', () => {
            if (fileInput) fileInput.click();
        });

        dropzone.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                if (fileInput) fileInput.click();
            }
        });

        let dragDepth = 0;
        dropzone.addEventListener('dragenter', (e) => {
            e.preventDefault();
            e.stopPropagation();
            dragDepth++;
            dropzone.classList.add('drag-over');
        });

        dropzone.addEventListener('dragover', (e) => {
            e.preventDefault();
            e.stopPropagation();
            e.dataTransfer.dropEffect = 'copy';
            dropzone.classList.add('drag-over');
        });

        dropzone.addEventListener('dragleave', (e) => {
            e.preventDefault();
            e.stopPropagation();
            dragDepth = Math.max(0, dragDepth - 1);
            if (dragDepth === 0) {
                dropzone.classList.remove('drag-over');
            }
        });

        dropzone.addEventListener('drop', (e) => {
            e.preventDefault();
            e.stopPropagation();
            dragDepth = 0;
            dropzone.classList.remove('drag-over');
            const dt = e.dataTransfer;
            if (dt && dt.files && dt.files.length > 0) {
                addFiles(dt.files);
            }
        });
    }

    const requiredElements = form.querySelectorAll('[required]');
    requiredElements.forEach(element => {
        element.addEventListener('invalid', function () {
            if (this.validity.valueMissing) {
                this.setCustomValidity('تکمیل این فیلد الزامی است.');
            }
        });
        element.addEventListener('input', function () { this.setCustomValidity(''); });
        element.addEventListener('change', function () { this.setCustomValidity(''); });
    });

    async function loadCaptcha(imgId, tokenInputId, answerInputId) {
        const imgEl = document.getElementById(imgId);
        if (imgEl) imgEl.classList.add('skeleton-shimmer');
        try {

            const res = await fetchWithTimeout('/api/captcha');
            if (res.ok) {
                const data = await res.json();
                const tokenEl = document.getElementById(tokenInputId);
                const answerEl = document.getElementById(answerInputId);

                if (imgEl) {
                    imgEl.src = data.image_base64;
                    imgEl.onload = () => imgEl.classList.remove('skeleton-shimmer');
                }
                if (tokenEl) tokenEl.value = data.captcha_id;
                if (answerEl) answerEl.value = '';
            } else {
                if (imgEl) imgEl.classList.remove('skeleton-shimmer');
            }
        } catch (e) {
            console.error('Failed to load CAPTCHA');
            if (imgEl) imgEl.classList.remove('skeleton-shimmer');
        }
    }

    document.getElementById('submit-captcha-refresh')?.addEventListener('click', () => {
        loadCaptcha('submit-captcha-img', 'submit_captcha_id', 'submit_captcha_answer');
    });

    document.getElementById('track-captcha-refresh')?.addEventListener('click', () => {
        loadCaptcha('track-captcha-img', 'track_captcha_id', 'track_captcha_answer');
    });

    function setBtnLoading(btn, isLoading, defaultText) {
        if (!btn) return;
        const btnText = btn.querySelector('.btn-text');
        const spinner = btn.querySelector('.btn-spinner');

        btn.disabled = isLoading;
        if (isLoading) {
            if (btnText) btnText.textContent = 'در حال رمزنگاری و ارسال...';
            if (spinner) spinner.classList.remove('d-none');
        } else {
            if (btnText) btnText.textContent = defaultText;
            if (spinner) spinner.classList.add('d-none');
        }
    }

    function checkClientVulgarLanguage(text) {
        if (!text || !text.trim()) return false;

        const lowered = text.toLowerCase()
            .replace(/ي/g, 'ی')
            .replace(/ئ/g, 'ی')
            .replace(/ك/g, 'ک')
            .replace(/ة/g, 'ه')
            .replace(/ؤ/g, 'و')
            .replace(/ـ/g, '')
            .replace(/[\u064B-\u065F\u0670\u200C\u200B\uFEFF\u200E\u200F]/g, '')
            .replace(/[@$013!45]/g, m => ({ '@': 'a', '$': 's', '0': 'o', '1': 'i', '3': 'e', '!': 'i', '4': 'a', '5': 's' })[m]);

        return _fastCheckClientVulgar(lowered);
    }

    const _vulgarPatterns = [
        /(?:^|[^\p{L}\p{N}])ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+ر(?:ی|م|ت|تون|مون|ش|ها)?(?:[^\p{L}\p{N}]|$)/iu,
        /ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[یiي*]+[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر/iu,
        /(?:^|[^\p{L}\p{N}])ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[صس][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+ش(?:[^\p{L}\p{N}]|$)/iu,
        /ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[وo0u*]?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[صس][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ک?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ش/iu,
        /(?:^|[^\p{L}\p{N}])ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+[صس](?:[^\p{L}\p{N}]|$)/iu,
        /ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[صس][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*خ[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*و?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ل/iu,
        /ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[صس][\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ش[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[ع*]?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر/iu,
        /(?:^|[^\p{L}\p{N}])ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+ن[یiي*]?(?:[^\p{L}\p{N}]|$)/iu,
        /ج[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ن?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هه‌a*]/iu,
        /ق[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ح[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ب?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هه‌a*]/iu,
        /د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*و+[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ث/iu,
        /س[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ک[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ت[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر/iu,
        /پ[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*س[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*گ/iu,
        /ب[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ن[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*م[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*و+[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*س/iu,
        /ب[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ش[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ف/iu,
        /ح[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[اوu]*[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*م[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ز[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هه‌a*]/iu,
        /م[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ر[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ج[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ن?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[هه‌a*]/iu,
        /گ[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ی+[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*د[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*م/iu,
        /(?:^|[^\p{L}\p{N}])ل[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ا?[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*ش[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]*[یiي*]+(?:[^\p{L}\p{N}]|$)/iu,
        /\bk+[i|e|y]+r+\b/i,
        /\bk+[o|u]+s+\b/i,
        /\bk+[o|u]+n+\b/i,
        /\bf+[\s*._~^#@$%&/\\|!?\-+=0-9]*[u*]+[\s*._~^#@$%&/\\|!?\-+=0-9]*[c*]*[\s*._~^#@$%&/\\|!?\-+=0-9]*k+\b/i,
        /\bs+[\s*._~^#@$%&/\\|!?\-+=0-9]*[h*]+[\s*._~^#@$%&/\\|!?\-+=0-9]*[i*!1]+[\s*._~^#@$%&/\\|!?\-+=0-9]*t+\b/i,
        /\bb+[\s*._~^#@$%&/\\|!?\-+=0-9]*[i*!1]+[\s*._~^#@$%&/\\|!?\-+=0-9]*t+[\s*._~^#@$%&/\\|!?\-+=0-9]*c+[\s*._~^#@$%&/\\|!?\-+=0-9]*h+\b/i,
        /\bc+[\s*._~^#@$%&/\\|!?\-+=0-9]*[u*]+[\s*._~^#@$%&/\\|!?\-+=0-9]*n+[\s*._~^#@$%&/\\|!?\-+=0-9]*t+\b/i,
        /\ba+[\s*._~^#@$%&/\\|!?\-+=0-9]*s+[\s*._~^#@$%&/\\|!?\-+=0-9]*s+[\s*._~^#@$%&/\\|!?\-+=0-9]*h+[\s*._~^#@$%&/\\|!?\-+=0-9]*o+[\s*._~^#@$%&/\\|!?\-+=0-9]*l+[\s*._~^#@$%&/\\|!?\-+=0-9]*e+\b/i,
        /\bj+e+n+d+e+h?/i,
        /\bd+a+y+[o|u]+s+/i,
        /\bk+o+s+k+e+s+h+/i
    ];

    const _stripSymbols = s => s.replace(/[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+/gu, '');
    const _collapseRepeats = s => s.replace(/(.)\1+/gu, '$1');
    const _toPhonetic = s => s.replace(/ص/g, 'س');

    const _rawPersianList = [
        "کیر", "کسکش", "کس کش", "کوسکش", "کوس کش", "کصکش", "کص کش", "کوصکش", "کوص کش",
        "کونی", "کوص", "کص", "کوس", "کون", "جنده", "قحبه",
        "مادرجنده", "مادر جنده", "خواهرجنده", "خواهر جنده", "پدر سگ", "پدرسگ", "پدرصلواتی",
        "دیوث", "بی ناموس", "بی‌ناموس", "بیناموس", "حرامزاده", "حرومزاده", "حروم زاده", "حرام زاده",
        "سیکتیر", "سیک تیر", "گوه", "عن", "لاشی", "لاشخور", "پفیوز", "کثافت", "بی شرف", "بی‌شرف",
        "بیشرف", "مادرقحبه", "مادر قحبه", "مادر قهوه", "مادرقهوه", "کیرم", "کیرت", "کیری",
        "کسخل", "کس خل", "کصخل", "کص خل", "کسمغز", "کس مغز", "کصمغز", "کص مغز",
        "کسشعر", "کس شعر", "کسشر", "کس شر", "کصشعر", "کص شعر", "کصشر", "کص شر", "کصش", "کسش",
        "گه نخور", "خفه شو", "کره خر", "تخم سگ", "تخمسگ", "گاایدم", "گاییدم", "میگام", "بگام",
        "ممه", "کوندوز", "کونده", "اوبی", "بی پدر", "بی‌پدر", "پدرسوخته", "پدر سوخته",
        "کسلیس", "کصلیس", "کونکش", "کون کش"
    ];

    const _precomputedPersianTokens = _rawPersianList.map(p => {
        const clean = _stripSymbols(p);
        return {
            raw: p,
            clean: clean,
            phonetic: _toPhonetic(clean),
            isPhrase: p.includes(' '),
            normPhrase: p.includes(' ') ? p.replace(/\s+/g, ' ') : null
        };
    });

    const _precomputedPhrases = _precomputedPersianTokens.filter(p => p.isPhrase);

    const _finglishList = [
        "kos", "koss", "kose", "koseh", "koskesh", "koskhol", "koskhel", "kir", "keer", "kiri", "kiram", "koon",
        "kooni", "jende", "jendeh", "madarjende", "khaharjende", "dayoos", "dayus", "dayous",
        "binamoos", "haroomzade", "haromzade", "siktir", "sikter", "lashi", "bisharaf",
        "pedarsag", "pedar sag", "goh", "gohkhor", "gayidam", "gayedam", "fuck", "shit", "bitch", "cunt", "asshole"
    ];

    const _vulgarSuffixes = ["", "ها", "هاش", "های", "ه", "ی", "م", "ت", "ش", "مان", "تان", "شان", "تون", "مون", "ان", "ا"];
    const _hasPrefixVulgar = (tok, term) => {
        if (tok.startsWith(term)) {
            const sfx = tok.substring(term.length);
            return _vulgarSuffixes.includes(sfx);
        }
        return false;
    };

    function _fastCheckClientVulgar(lowered) {
        if (_vulgarPatterns.some(rx => rx.test(lowered))) return true;

        const tokens = lowered.split(/\s+/);
        let singleCharSeq = '';

        for (let i = 0; i < tokens.length; i++) {
            const tok = tokens[i];
            const stripped = _stripSymbols(tok);
            const collapsed = _collapseRepeats(stripped);
            const phonetic = _toPhonetic(collapsed);

            for (let j = 0; j < _precomputedPersianTokens.length; j++) {
                const p = _precomputedPersianTokens[j];
                const pClean = p.clean;
                const pPhonetic = p.phonetic;

                if (stripped === pClean || collapsed === pClean || phonetic === pPhonetic) return true;
                if (_hasPrefixVulgar(stripped, pClean) || _hasPrefixVulgar(collapsed, pClean) || _hasPrefixVulgar(phonetic, pPhonetic)) return true;
            }

            for (let k = 0; k < _finglishList.length; k++) {
                const f = _finglishList[k];
                if (stripped === f || collapsed === f) return true;
                if (f.length >= 4 && (stripped.startsWith(f) || collapsed.startsWith(f))) return true;
            }

            if (stripped.length === 1) {
                singleCharSeq += stripped;
            } else if (singleCharSeq.length > 0) {
                const seqCol = _collapseRepeats(singleCharSeq);
                const seqPhon = _toPhonetic(seqCol);
                for (let j = 0; j < _precomputedPersianTokens.length; j++) {
                    const p = _precomputedPersianTokens[j];
                    if (singleCharSeq === p.clean || seqCol === p.clean || seqPhon === p.phonetic) return true;
                }
                singleCharSeq = '';
            }
        }

        if (singleCharSeq.length > 0) {
            const seqCol = _collapseRepeats(singleCharSeq);
            const seqPhon = _toPhonetic(seqCol);
            for (let j = 0; j < _precomputedPersianTokens.length; j++) {
                const p = _precomputedPersianTokens[j];
                if (singleCharSeq === p.clean || seqCol === p.clean || seqPhon === p.phonetic) return true;
            }
        }

        const normalizedSentence = lowered.replace(/[\s*._~^#@$%&/\\|!?\-+=×÷·•…—–]+/gu, ' ').trim();
        for (let m = 0; m < _precomputedPhrases.length; m++) {
            if (normalizedSentence.includes(_precomputedPhrases[m].normPhrase)) return true;
        }

        return false;
    }

    const LEGAL_WARNING_HTML = `
        <div class="legal-warning-box">
            <div class="legal-warning-header">
                <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"></path>
                    <line x1="12" y1="9" x2="12" y2="13"></line>
                    <line x1="12" y1="17" x2="12.01" y2="17"></line>
                </svg>
                <span>اخطار حقوقی و امنیتی بسیار مهم</span>
            </div>
            <div class="legal-warning-body">
                متن ارسالی شما حاوی الفاظ نامتعارف، توهین‌آمیز یا مصادیق توهین به اشخاص و کارکنان دولت می‌باشد. سامانه از پذیرش این پیام معذور است.<br>
                بر اساس <strong>ماده ۶۰۸ و ۶۰۹ قانون مجازات اسلامی</strong> و مقررات جرایم رایانه‌ای، هرگونه توهین، افترا و هتک حیثیت در سامانه‌های رسمی کشور، پیگرد قانونی به همراه داشته و مشخصات فنی مرتکب جهت رسیدگی به <strong>مراجع ذی‌صلاح قضایی و امنیتی</strong> منعکس خواهد گردید. لطفاً گزارش خود را با رعایت موازین قانونی و اداری بازنویسی فرمایید.
            </div>
        </div>
    `;

    function showConfirmDialog() {
        const category = document.getElementById('category');
        const desc = descTextarea?.value || '';
        const descPreview = desc.length > 120 ? desc.substring(0, 120) + '...' : desc;

        let attachmentSummary = 'ندارد';
        if (selectedFiles.length === 1) {
            attachmentSummary = `${escapeHtml(selectedFiles[0].name)} (${formatFileSize(selectedFiles[0].size)})`;
        } else if (selectedFiles.length > 1) {
            const totalBytes = selectedFiles.reduce((s, f) => s + f.size, 0);
            const names = selectedFiles.map(f => escapeHtml(f.name)).join('، ');
            attachmentSummary = `${toPersianNum(selectedFiles.length)} فایل (مجموعاً ${formatFileSize(totalBytes)}): ${names}`;
        }

        const summaryHtml = `
            <div class="confirm-summary">
                <div class="confirm-summary-item">
                    <span class="confirm-summary-label">موضوع:</span>
                    <span class="confirm-summary-value">${escapeHtml(category?.options[category.selectedIndex]?.text || 'نامشخص')}</span>
                </div>
                <div class="confirm-summary-item">
                    <span class="confirm-summary-label">شرح گزارش:</span>
                    <span class="confirm-summary-value">${escapeHtml(descPreview || 'بدون شرح')}</span>
                </div>
                <div class="confirm-summary-item">
                    <span class="confirm-summary-label">فایل‌های پیوست:</span>
                    <span class="confirm-summary-value">${attachmentSummary}</span>
                </div>
            </div>
        `;

        if (typeof window.showConfirmDialog === 'function') {
            return window.showConfirmDialog({
                title: 'بررسی نهایی قبل از ارسال گزارش',
                description: 'لطفاً پیش از ارسال نهایی، مشخصات گزارش محرمانه خود را به دقت بررسی فرمایید.',
                htmlContent: summaryHtml,
                confirmText: 'تأیید و ارسال گزارش',
                cancelText: 'بازگشت و ویرایش',
                variant: 'primary',
                iconType: 'check-circle'
            });
        }
        return Promise.resolve(window.confirm('آیا از ارسال گزارش اطمینان دارید؟'));
    }

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        if (!validateCurrentStep()) return;

        const descTrimmed = descTextarea?.value?.trim() || '';
        if (descTrimmed.length < MIN_REPORT_DESC_LENGTH) {
            currentStep = 2;
            updateStepTracker();
            descTextarea.focus();
            descTextarea.classList.add('error-pulse');
            setTimeout(() => descTextarea.classList.remove('error-pulse'), 1200);
            showToast(`متن شرح گزارش بسیار کوتاه است؛ لطفاً حداقل ۵۰ کاراکتر شرح و جزئیات وارد فرمایید (تعداد فعلی: ${toPersianNum(descTrimmed.length)} کاراکتر).`, 'warning');
            return;
        }

        const descVal = descTextarea?.value || '';
        const nameVal = document.getElementById('reporter_name')?.value || '';

        if (checkClientVulgarLanguage(descVal) || checkClientVulgarLanguage(nameVal)) {
            responseBox.innerHTML = LEGAL_WARNING_HTML;
            responseBox.style.display = 'block';
            responseBox.scrollIntoView({ behavior: getScrollBehavior(), block: 'center' });
            loadCaptcha('submit-captcha-img', 'submit_captcha_id', 'submit_captcha_answer');
            return;
        }

        const confirmed = await showConfirmDialog();
        if (!confirmed) return;

        setBtnLoading(submitBtn, true, 'ثبت گزارش');
        responseBox.style.display = 'none';
        responseBox.className = 'response-box';

        const categorySelectEl = document.getElementById('category');
        const submittedReceiptMeta = {
            category: categorySelectEl?.options[categorySelectEl.selectedIndex]?.text?.trim() || 'عمومی و سایر موضوعات',
            reporterName: document.getElementById('reporter_name')?.value?.trim() || '',
            reporterPhone: document.getElementById('reporter_phone')?.value?.trim() || '',
            fileCount: selectedFiles.length,
            submissionDate: new Date()
        };

        const formData = new FormData(form);
        formData.delete('evidence');
        for (const file of selectedFiles) {
            formData.append('evidence', file, file.name);
        }

        try {

            const response = await fetchWithTimeout('/api/report', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const errorText = await response.text();
                if (errorText.includes('اخطار حقوقی و امنیتی') || errorText.includes('ماده ۶۰۸')) {
                    responseBox.innerHTML = LEGAL_WARNING_HTML;
                    responseBox.style.display = 'block';
                    responseBox.scrollIntoView({ behavior: getScrollBehavior(), block: 'center' });
                    loadCaptcha('submit-captcha-img', 'submit_captcha_id', 'submit_captcha_answer');
                    return;
                }
                throw new Error(errorText || 'خطا در برقراری ارتباط با سرور.');
            }

            const data = await response.json();

            if (data.status === 'success') {

                populateOfficialReceipt({
                    tracking_code: data.tracking_code,
                    category: submittedReceiptMeta.category,
                    type: (submittedReceiptMeta.reporterName || submittedReceiptMeta.reporterPhone) ? 'با هویت (محرمانه)' : 'ناشناس (محرمانه)',
                    reporter: submittedReceiptMeta.reporterName || 'محفوظ (ثبت ناشناس)',
                    phone: submittedReceiptMeta.reporterPhone ? toPersianNum(submittedReceiptMeta.reporterPhone) : 'ثبت نشده',
                    attachments: submittedReceiptMeta.fileCount > 0 ? `${toPersianNum(submittedReceiptMeta.fileCount)} فایل پیوست` : 'فاقد فایل پیوست',
                    timestamp: submittedReceiptMeta.submissionDate
                });

                form.classList.add('fade-out-anim');

                setTimeout(() => {
                    form.classList.add('d-none');
                    form.classList.remove('fade-out-anim');

                    const successScreen = document.getElementById('success-screen');
                    const trackingEl = document.getElementById('final-tracking-code');
                    if (trackingEl) trackingEl.textContent = data.tracking_code;
                    successScreen?.classList.remove('d-none');

                    form.reset();
                    clearFileInput();
                    currentStep = 1;
                    updateStepTracker();
                    clearDraft();
                    loadCaptcha('submit-captcha-img', 'submit_captcha_id', 'submit_captcha_answer');
                    showToast('گزارش شما با موفقیت ثبت و رمزنگاری شد.', 'success', 5000);
                }, 250);
            } else {
                throw new Error('پاسخ نامعتبر از سرور.');
            }

        } catch (error) {
            responseBox.classList.add('error');
            responseBox.textContent = `خطا: ${error.message}`;
            responseBox.style.display = 'block';
            loadCaptcha('submit-captcha-img', 'submit_captcha_id', 'submit_captcha_answer');
        } finally {
            setBtnLoading(submitBtn, false, 'ثبت گزارش');
        }
    });

    const copyTrackingBtn = document.getElementById('copyTrackingBtn');
    const copyBtnText = document.getElementById('copyBtnText');
    const finalTrackingCode = document.getElementById('final-tracking-code');

    if (copyTrackingBtn && finalTrackingCode) {
        copyTrackingBtn.addEventListener('click', async () => {
            const code = finalTrackingCode.textContent.trim();
            if (!code) return;

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

                copyTrackingBtn.classList.add('copied');
                if (copyBtnText) copyBtnText.textContent = 'کپی شد!';

                setTimeout(() => {
                    copyTrackingBtn.classList.remove('copied');
                    if (copyBtnText) copyBtnText.textContent = 'کپی کد';
                }, 2000);
            } catch (err) {
                console.error('Failed to copy text: ', err);
            }
        });
    }

    const goToTrackingLink = document.getElementById('goToTrackingLink');
    if (goToTrackingLink) {
        goToTrackingLink.addEventListener('click', (e) => {
            e.preventDefault();
            const trackCodeInput = document.getElementById('track_code_input');
            if (finalTrackingCode && trackCodeInput) {
                const code = finalTrackingCode.textContent.trim();
                if (code) {
                    trackCodeInput.value = code;
                    trackCodeInput.dispatchEvent(new Event('input', { bubbles: true }));
                }
            }
            tabTrack?.click();
            window.scrollTo({ top: 0, behavior: getScrollBehavior() });
        });
    }

    const pasteTrackingBtn = document.getElementById('pasteTrackingBtn');
    const pasteBtnText = document.getElementById('pasteBtnText');
    const trackCodeInput = document.getElementById('track_code_input');

    if (pasteTrackingBtn && trackCodeInput) {
        pasteTrackingBtn.addEventListener('click', async () => {
            try {
                if (navigator.clipboard && navigator.clipboard.readText) {
                    const clipText = await navigator.clipboard.readText();
                    if (clipText && clipText.trim()) {
                        const cleaned = clipText.trim().toUpperCase();
                        trackCodeInput.value = cleaned;
                        trackCodeInput.dispatchEvent(new Event('input', { bubbles: true }));
                        trackCodeInput.focus();

                        pasteTrackingBtn.classList.add('pasted');
                        if (pasteBtnText) pasteBtnText.textContent = 'جای‌گذاری شد!';

                        setTimeout(() => {
                            pasteTrackingBtn.classList.remove('pasted');
                            if (pasteBtnText) pasteBtnText.textContent = 'جای‌گذاری سریع';
                        }, 2000);
                    }
                } else {
                    trackCodeInput.focus();
                    trackCodeInput.select();
                }
            } catch (err) {
                console.warn('Clipboard read error or permission denied:', err);
                trackCodeInput.focus();
            }
        });
    }

    document.getElementById('printReceiptBtn')?.addEventListener('click', () => {
        const activeTracking = document.getElementById('final-tracking-code')?.textContent?.trim();
        const receiptTracking = document.getElementById('receipt-tracking-code')?.textContent?.trim();
        if (activeTracking && (!receiptTracking || receiptTracking === 'CZRLUUTM2XKMNCD3')) {
            populateOfficialReceipt({
                tracking_code: activeTracking,
                category: 'عمومی و سایر موضوعات',
                type: 'محرمانه',
                reporter: 'محفوظ (ثبت ناشناس)',
                phone: 'ثبت نشده',
                attachments: 'فاقد فایل پیوست'
            });
        }
        window.print();
    });

    document.getElementById('newReportBtn')?.addEventListener('click', () => {
        const successScreen = document.getElementById('success-screen');
        successScreen?.classList.add('fade-out-anim');

        setTimeout(() => {
            successScreen?.classList.add('d-none');
            successScreen?.classList.remove('fade-out-anim');

            form.classList.remove('d-none');
            responseBox.style.display = 'none';
            loadCaptcha('submit-captcha-img', 'submit_captcha_id', 'submit_captcha_answer');
            window.scrollTo({ top: 0, behavior: getScrollBehavior() });
        }, 250);
    });

    const tabSubmit = document.getElementById('tab-submit');
    const tabTrack = document.getElementById('tab-track');
    const submitSection = document.getElementById('submitSection');
    const trackingSection = document.getElementById('trackingSection');

    if (tabSubmit && tabTrack) {
        tabSubmit.addEventListener('click', () => {
            tabSubmit.classList.add('active');
            tabTrack.classList.remove('active');
            submitSection?.classList.remove('d-none');
            trackingSection?.classList.add('d-none');

            const successScreen = document.getElementById('success-screen');
            const isSuccessVisible = successScreen && !successScreen.classList.contains('d-none');
            if (!isSuccessVisible) {
                loadCaptcha('submit-captcha-img', 'submit_captcha_id', 'submit_captcha_answer');
            }
        });

        tabTrack.addEventListener('click', () => {
            tabTrack.classList.add('active');
            tabSubmit.classList.remove('active');
            trackingSection?.classList.remove('d-none');
            submitSection?.classList.add('d-none');
            loadCaptcha('track-captcha-img', 'track_captcha_id', 'track_captcha_answer');
        });
    }

    const openDrawerBtn = document.getElementById('openDrawerBtn');
    const closeDrawerBtn = document.getElementById('closeDrawerBtn');
    const guidelinesDrawer = document.getElementById('guidelinesDrawer');

    function openGuidelinesDrawer() {
        window.scrollTo({ top: 0, behavior: getScrollBehavior() });

        if (guidelinesDrawer) {
            guidelinesDrawer.classList.remove('d-none');
            document.body.style.overflow = 'hidden';
            const panel = guidelinesDrawer.querySelector('.drawer-panel');
            if (panel) {

                promoteTemporarily(panel, 'transform', 350);
                panel.scrollTop = 0;
            }
        }
    }

    function closeGuidelinesDrawer() {
        if (guidelinesDrawer) {
            const panel = guidelinesDrawer.querySelector('.drawer-panel');
            if (panel) {

                promoteTemporarily(panel, 'transform', 350);
            }
            guidelinesDrawer.classList.add('d-none');
            document.body.style.overflow = '';
        }
    }

    if (openDrawerBtn) openDrawerBtn.addEventListener('click', openGuidelinesDrawer);
    if (closeDrawerBtn) closeDrawerBtn.addEventListener('click', closeGuidelinesDrawer);

    if (guidelinesDrawer) {
        guidelinesDrawer.addEventListener('click', (e) => {
            if (e.target === guidelinesDrawer) closeGuidelinesDrawer();
        });
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && !guidelinesDrawer.classList.contains('d-none')) {
                closeGuidelinesDrawer();
            }
        });
    }

    const faqItems = document.querySelectorAll('.faq-accordion-item');
    faqItems.forEach((item) => {
        const header = item.querySelector('.faq-accordion-header');
        if (header) {
            header.addEventListener('click', () => {
                const isOpen = item.classList.contains('open');

                faqItems.forEach((other) => {
                    if (other !== item) {
                        other.classList.remove('open');
                        other.querySelector('.faq-accordion-header')?.setAttribute('aria-expanded', 'false');
                    }
                });

                item.classList.toggle('open', !isOpen);
                header.setAttribute('aria-expanded', !isOpen ? 'true' : 'false');
            });
        }
    });

    document.getElementById('followup-captcha-refresh')?.addEventListener('click', () => {
        loadCaptcha('followup-captcha-img', 'followup_captcha_id', 'followup_captcha_answer');
    });

    function renderConversationThread(reportData) {
        const threadContainer = document.getElementById('track-conversation-container');
        const messagesList = document.getElementById('track-messages-list');
        const countBadge = document.getElementById('track-messages-count-badge');
        if (!threadContainer || !messagesList) return;

        messagesList.innerHTML = '';

        const allItems = [];

        if (reportData && reportData.admin_response && reportData.admin_response.trim() !== '') {
            allItems.push({
                sender_type: 'admin',
                is_official_response: true,
                sender_title: 'پاسخ و نتیجه بررسی اداره حراست',
                message: reportData.admin_response.trim(),
                created_at: reportData.responded_at || reportData.created_at
            });
        }

        if (reportData && Array.isArray(reportData.messages)) {
            reportData.messages.forEach(msg => {
                const isRep = msg.sender_type === 'reporter';
                allItems.push({
                    sender_type: msg.sender_type,
                    is_official_response: false,
                    sender_title: isRep ? 'پیام تکمیلی شما' : 'پاسخ اداره حراست',
                    message: msg.message,
                    created_at: msg.created_at
                });
            });
        }

        allItems.sort((a, b) => new Date(a.created_at || 0) - new Date(b.created_at || 0));

        if (countBadge) {
            countBadge.textContent = allItems.length > 0 ? `${toPersianNum(allItems.length)} مورد ثبت‌شده` : 'فاقد پاسخ یا پیام تکمیلی';
        }

        threadContainer.classList.remove('d-none');

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
                        <div class="nexus-bubble-avatar ${avatarClass}" title="${escapeHtml(item.sender_title)}">
                            ${avatarSvg}
                        </div>
                        <div class="nexus-bubble-card ${bubbleCardClass}">
                            <div class="nexus-bubble-top">
                                <span class="nexus-bubble-sender">${escapeHtml(item.sender_title)}</span>
                                ${timeStr ? `<span class="nexus-bubble-timestamp">${timeStr}</span>` : ''}
                            </div>
                            <div class="nexus-bubble-body">${escapeHtml(item.message).replace(/\n/g, '<br>')}</div>
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
                <span class="nexus-chat-empty-title">هنوز پاسخی برای این گزارش ثبت نشده است</span>
                <span class="nexus-chat-empty-sub">گزارش شما در کارتابل امن حراست قرار دارد. به محض بررسی کارشناسان، نتیجه و پاسخ رسمی در این کارت به صورت گفتگو ثبت می‌گردد. همچنین در صورت تمایل می‌توانید از کادر زیر، توضیحات تکمیلی جدید ارسال فرمایید.</span>
            `;
            messagesList.appendChild(emptyCard);
        }
    }

    const btnSubmitFollowup = document.getElementById('btnSubmitFollowup');
    const followupInput = document.getElementById('followup-message-input');
    const followupStatusMsg = document.getElementById('followup-status-msg');

    if (btnSubmitFollowup) {
        btnSubmitFollowup.addEventListener('click', async () => {
            const code = trackInput.value.trim().toUpperCase();
            const message = followupInput?.value.trim();
            const captchaId = document.getElementById('followup_captcha_id')?.value;
            const captchaAns = document.getElementById('followup_captcha_answer')?.value;

            if (!code) {
                showToast('کد پیگیری معتبر نیست.', 'warning');
                return;
            }
            if (!message) {
                showToast('لطفاً متن پیام تکمیلی را وارد کنید.', 'warning');
                return;
            }

            if (checkClientVulgarLanguage(message)) {
                if (followupStatusMsg) {
                    followupStatusMsg.innerHTML = LEGAL_WARNING_HTML;
                    followupStatusMsg.style.display = 'block';
                }
                loadCaptcha('followup-captcha-img', 'followup_captcha_id', 'followup_captcha_answer');
                return;
            }

            if (!captchaAns) {
                showToast('لطفاً کد امنیتی را وارد کنید.', 'warning');
                return;
            }

            setBtnLoading(btnSubmitFollowup, true, 'ارسال پیام تکمیلی');
            if (followupStatusMsg) followupStatusMsg.style.display = 'none';

            const messagesList = document.getElementById('track-messages-list');
            const threadContainer = document.getElementById('track-conversation-container');
            const countBadge = document.getElementById('track-messages-count-badge');

            let optimisticBubble = null;
            if (messagesList && threadContainer) {
                threadContainer.classList.remove('d-none');
                const emptyCard = messagesList.querySelector('.nexus-chat-empty');
                if (emptyCard) emptyCard.remove();

                const userSvg = '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>';
                optimisticBubble = document.createElement('div');
                optimisticBubble.className = 'nexus-bubble-row nexus-row-reporter';
                optimisticBubble.id = 'optimistic-followup-row';
                optimisticBubble.innerHTML = `
                    <div class="nexus-bubble-unit">
                        <div class="nexus-bubble-avatar nexus-avatar-reporter" style="opacity: 0.7;">
                            ${userSvg}
                        </div>
                        <div class="nexus-bubble-card nexus-bubble-reporter" style="opacity: 0.82; border-style: dashed;">
                            <div class="nexus-bubble-top">
                                <span class="nexus-bubble-sender">پیام تکمیلی شما (در حال ارسال...)</span>
                                <span class="nexus-bubble-timestamp" style="display:inline-flex; align-items:center; gap:4px;">
                                    <span class="btn-spinner" style="width:10px; height:10px; display:inline-block; border-width:1.5px; border-color: rgba(255,255,255,0.3) transparent transparent transparent;"></span>
                                    ارسال
                                </span>
                            </div>
                            <div class="nexus-bubble-body">${escapeHtml(message).replace(/\n/g, '<br>')}</div>
                        </div>
                    </div>
                `;
                messagesList.appendChild(optimisticBubble);
                optimisticBubble.scrollIntoView({ behavior: getScrollBehavior(), block: 'nearest' });
            }

            try {

                const res = await fetchWithTimeout('/api/report/follow-up', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        tracking_code: code,
                        message: message,
                        captcha_id: captchaId,
                        captcha_answer: captchaAns
                    })
                });

                if (optimisticBubble && optimisticBubble.parentNode) {
                    optimisticBubble.remove();
                }

                if (!res.ok) {
                    const errText = await res.text();
                    if (errText.includes('اخطار حقوقی و امنیتی') || errText.includes('ماده ۶۰۸')) {
                        if (followupStatusMsg) {
                            followupStatusMsg.innerHTML = LEGAL_WARNING_HTML;
                            followupStatusMsg.style.display = 'block';
                        }
                        loadCaptcha('followup-captcha-img', 'followup_captcha_id', 'followup_captcha_answer');
                        return;
                    }
                    throw new Error(errText || 'خطا در ثبت پیام تکمیلی');
                }

                const data = await res.json();
                if (followupStatusMsg) {
                    followupStatusMsg.className = 'response-box success';
                    followupStatusMsg.textContent = data.message || 'پیام تکمیلی با موفقیت ثبت شد.';
                    followupStatusMsg.style.display = 'block';
                }

                if (followupInput) followupInput.value = '';
                document.getElementById('followup_captcha_answer').value = '';
                loadCaptcha('followup-captcha-img', 'followup_captcha_id', 'followup_captcha_answer');

                if (messagesList && threadContainer) {
                    const userSvg = '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>';
                    const row = document.createElement('div');
                    row.className = 'nexus-bubble-row nexus-row-reporter';
                    row.innerHTML = `
                        <div class="nexus-bubble-unit">
                            <div class="nexus-bubble-avatar nexus-avatar-reporter" title="پیام تکمیلی شما">
                                ${userSvg}
                            </div>
                            <div class="nexus-bubble-card nexus-bubble-reporter">
                                <div class="nexus-bubble-top">
                                    <span class="nexus-bubble-sender">پیام تکمیلی شما</span>
                                    <span class="nexus-bubble-timestamp">هم‌اکنون</span>
                                </div>
                                <div class="nexus-bubble-body">${escapeHtml(message).replace(/\n/g, '<br>')}</div>
                            </div>
                        </div>
                    `;
                    messagesList.appendChild(row);
                    row.scrollIntoView({ behavior: getScrollBehavior(), block: 'nearest' });

                    if (countBadge) {
                        const currentCount = messagesList.querySelectorAll('.nexus-bubble-row').length;
                        countBadge.textContent = `${toPersianNum(currentCount)} مورد ثبت‌شده`;
                    }
                }
            } catch (err) {
                if (optimisticBubble && optimisticBubble.parentNode) {
                    optimisticBubble.remove();
                }
                if (followupStatusMsg) {
                    followupStatusMsg.className = 'response-box error';
                    followupStatusMsg.textContent = err.message;
                    followupStatusMsg.style.display = 'block';
                }
                loadCaptcha('followup-captcha-img', 'followup_captcha_id', 'followup_captcha_answer');
            } finally {
                setBtnLoading(btnSubmitFollowup, false, 'ارسال پیام تکمیلی');
            }
        });
    }

    const trackForm = document.getElementById('trackForm');
    const trackInput = document.getElementById('track_code_input');
    const trackCaptchaAnswer = document.getElementById('track_captcha_answer');
    const trackCaptchaId = document.getElementById('track_captcha_id');
    const trackSubmitBtn = document.getElementById('trackSubmitBtn');
    const trackResultBox = document.getElementById('trackResultBox');
    const trackResultSkeleton = document.getElementById('trackResultSkeleton');
    const trackErrorBox = document.getElementById('trackErrorBox');

    function showTrackError(msg) {
        if (!trackErrorBox) return;
        trackErrorBox.className = 'response-box error';
        trackErrorBox.innerHTML = `
            <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"></circle>
                <line x1="12" y1="8" x2="12" y2="12"></line>
                <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>
            <span>${escapeHtml(msg)}</span>
        `;
        trackErrorBox.style.display = 'flex';
        trackErrorBox.scrollIntoView({ behavior: getScrollBehavior(), block: 'nearest' });
    }

    trackInput?.addEventListener('input', () => {
        if (trackErrorBox && trackErrorBox.style.display !== 'none') {
            trackErrorBox.style.display = 'none';
        }
    });

    trackCaptchaAnswer?.addEventListener('input', () => {
        if (trackErrorBox && trackErrorBox.style.display !== 'none') {
            trackErrorBox.style.display = 'none';
        }
    });

    if (trackForm) {
        trackForm.addEventListener('submit', async (e) => {
            e.preventDefault();

            const code = trackInput.value.trim().toUpperCase();
            if (!code) {
                showTrackError('لطفاً کد پیگیری ۱۶ رقمی گزارش را وارد نمایید.');
                trackResultBox?.classList.add('d-none');
                const inputContainer = trackInput.closest('.tracking-input-box') || trackInput;
                inputContainer.classList.add('error-pulse');
                setTimeout(() => inputContainer.classList.remove('error-pulse'), 1200);
                trackInput.focus();
                showToast('لطفاً کد پیگیری ۱۶ رقمی گزارش را وارد نمایید.', 'warning');
                return;
            }

            const captchaAns = trackCaptchaAnswer?.value?.trim() || '';
            if (!captchaAns) {
                showTrackError('لطفاً کد امنیتی را وارد نمایید.');
                trackResultBox?.classList.add('d-none');
                if (trackCaptchaAnswer) {
                    trackCaptchaAnswer.classList.add('error-pulse');
                    setTimeout(() => trackCaptchaAnswer.classList.remove('error-pulse'), 1200);
                    trackCaptchaAnswer.focus();
                }
                showToast('لطفاً کد امنیتی را وارد نمایید.', 'warning');
                return;
            }

            setBtnLoading(trackSubmitBtn, true, 'استعلام وضعیت و مشاهده پاسخ');
            trackErrorBox.style.display = 'none';
            trackResultBox?.classList.add('d-none');

            let skeletonTimer = setTimeout(() => {
                if (trackResultSkeleton) {
                    trackResultSkeleton.classList.remove('d-none');
                }
            }, 80);

            try {

                const res = await fetchWithTimeout('/api/report/track', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        tracking_code: code,
                        captcha_id: trackCaptchaId.value,
                        captcha_answer: captchaAns
                    })
                });

                clearTimeout(skeletonTimer);
                if (trackResultSkeleton) {
                    trackResultSkeleton.classList.add('d-none');
                }

                if (!res.ok) {
                    const errTxt = await res.text();
                    if (res.status === 404) {
                        throw new Error('گزارشی با این کد پیگیری در سامانه یافت نشد.');
                    }
                    throw new Error(errTxt || 'خطا در برقراری ارتباط با سامانه.');
                }

                const data = await res.json();

                let shamsiDate = formatJalaliDateTime(data.created_at);

                document.getElementById('track-category').textContent = data.category || 'عمومی';
                document.getElementById('track-date').textContent = shamsiDate;

                const statusEl = document.getElementById('track-status');
                if (statusEl) {
                    statusEl.innerHTML = '';
                    const badge = document.createElement('span');
                    const hasAdminMsg = Array.isArray(data.messages) && data.messages.some(m => m.sender_type === 'admin');
                    const hasResponse = Boolean(
                        (data.admin_response && data.admin_response.trim() !== '') ||
                        data.has_response ||
                        data.responded_at ||
                        hasAdminMsg
                    );
                    if (hasResponse) {
                        badge.className = 'badge-status responded';
                        badge.innerHTML = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0;"><polyline points="20 6 9 17 4 12"></polyline></svg><span>پاسخ داده شده</span>`;
                    } else if (data.is_read) {
                        badge.className = 'badge-status read';
                        badge.innerHTML = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0;"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"></path><circle cx="12" cy="12" r="3"></circle></svg><span>بررسی شده</span>`;
                    } else {
                        badge.className = 'badge-status unread';
                        badge.innerHTML = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0;"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg><span>در نوبت بررسی</span>`;
                    }
                    statusEl.appendChild(badge);
                }

                renderConversationThread(data);
                loadCaptcha('followup-captcha-img', 'followup_captcha_id', 'followup_captcha_answer');

                trackResultBox?.classList.remove('d-none');
                requestAnimationFrame(() => {
                    setTimeout(() => {
                        trackResultBox?.scrollIntoView({
                            behavior: getScrollBehavior(),
                            block: 'start'
                        });
                    }, 50);
                });
            } catch (err) {
                clearTimeout(skeletonTimer);
                if (trackResultSkeleton) {
                    trackResultSkeleton.classList.add('d-none');
                }
                showTrackError(err.message);
                loadCaptcha('track-captcha-img', 'track_captcha_id', 'track_captcha_answer');
            } finally {
                setBtnLoading(trackSubmitBtn, false, 'استعلام وضعیت و مشاهده پاسخ');
            }
        });
    }
});
