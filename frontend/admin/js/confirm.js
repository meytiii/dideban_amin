(function () {
    'use strict';

    function safeEscape(str) {
        if (!str) return '';
        return String(str).replace(/[&<>'"]/g, tag => ({
            '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;'
        }[tag] || tag));
    }

    const icons = {
        'logout': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>',
        'trash': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 10V17M8 10V17M19 9H22M19 14H22M19 19H21M16 6V16.2C16 17.8802 16 18.7202 15.673 19.362C15.3854 19.9265 14.9265 20.3854 14.362 20.673C13.7202 21 12.8802 21 11.2 21H8.8C7.11984 21 6.27976 21 5.63803 20.673C5.07354 20.3854 4.6146 19.9265 4.32698 19.362C4 18.7202 4 17.8802 4 16.2V6M2 6H18M14 6L13.7294 5.18807C13.4671 4.40125 13.3359 4.00784 13.0927 3.71698C12.8779 3.46013 12.6021 3.26132 12.2905 3.13878C11.9376 3 11.523 3 10.6936 3H9.30643C8.47705 3 8.06236 3 7.70951 3.13878C7.39792 3.26132 7.12208 3.46013 6.90729 3.71698C6.66405 4.00784 6.53292 4.40125 6.27064 5.18807L6 6"></path></svg>',
        'shield-alert': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none"><path d="M12 7.25C12.4142 7.25 12.75 7.58579 12.75 8V13C12.75 13.4142 12.4142 13.75 12 13.75C11.5858 13.75 11.25 13.4142 11.25 13V8C11.25 7.58579 11.5858 7.25 12 7.25Z" fill="currentColor"/><path d="M12 17C12.5523 17 13 16.5523 13 16C13 15.4477 12.5523 15 12 15C11.4477 15 11 15.4477 11 16C11 16.5523 11.4477 17 12 17Z" fill="currentColor"/><path fill-rule="evenodd" clip-rule="evenodd" d="M8.2944 4.47643C9.36631 3.11493 10.5018 2.25 12 2.25C13.4981 2.25 14.6336 3.11493 15.7056 4.47643C16.7598 5.81544 17.8769 7.79622 19.3063 10.3305L19.7418 11.1027C20.9234 13.1976 21.8566 14.8523 22.3468 16.1804C22.8478 17.5376 22.9668 18.7699 22.209 19.8569C21.4736 20.9118 20.2466 21.3434 18.6991 21.5471C17.1576 21.75 15.0845 21.75 12.4248 21.75H11.5752C8.91552 21.75 6.84239 21.75 5.30082 21.5471C3.75331 21.3434 2.52637 20.9118 1.79099 19.8569C1.03318 18.7699 1.15218 17.5376 1.65314 16.1804C2.14334 14.8523 3.07658 13.1977 4.25818 11.1027L4.69361 10.3307C6.123 7.79629 7.24019 5.81547 8.2944 4.47643ZM9.47297 5.40432C8.49896 6.64148 7.43704 8.51988 5.96495 11.1299L5.60129 11.7747C4.37507 13.9488 3.50368 15.4986 3.06034 16.6998C2.6227 17.8855 2.68338 18.5141 3.02148 18.9991C3.38202 19.5163 4.05873 19.8706 5.49659 20.0599C6.92858 20.2484 8.9026 20.25 11.6363 20.25H12.3636C15.0974 20.25 17.0714 20.2484 18.5034 20.0599C19.9412 19.8706 20.6179 19.5163 20.9785 18.9991C21.3166 18.5141 21.3773 17.8855 20.9396 16.6998C20.4963 15.4986 19.6249 13.9488 18.3987 11.7747L18.035 11.1299C16.5629 8.51987 15.501 6.64148 14.527 5.40431C13.562 4.17865 12.8126 3.75 12 3.75C11.1874 3.75 10.4379 4.17865 9.47297 5.40432Z" fill="currentColor"/></svg>',
        'shield-question': '<svg viewBox="0 0 31.694 31.694" width="24" height="24" fill="currentColor"><path d="M23.112,9.35c-4.473,2.742-7.697,6.205-9.006,7.744l-3.361-2.633c-0.089-0.064-1.934,1.574-1.885,1.625l6.121,6.223 c0.045,0.051,0.107,0.074,0.174,0.074c0.012,0,0.031,0,0.047-0.004c0.08-0.018,0.148-0.072,0.178-0.152 c0.986-2.521,4.242-7.799,8.4-11.627c0.074-0.07,0.1-0.18,0.059-0.275C23.838,10.324,23.174,9.312,23.112,9.35z"/><path d="M27.126,3.842c-9.268,0-10.836-3.518-10.85-3.551C16.208,0.119,16.04,0.004,15.85,0c0,0-0.004,0-0.008,0 c-0.184,0-0.354,0.115-0.428,0.283C15.403,0.32,13.801,3.842,4.568,3.842c-0.258,0-0.462,0.209-0.462,0.461v15.764 c0,6.453,11.084,11.383,11.553,11.59c0.062,0.027,0.121,0.037,0.188,0.037c0.061,0,0.127-0.01,0.186-0.037 c0.473-0.207,11.555-5.137,11.555-11.59V4.303C27.586,4.051,27.381,3.842,27.126,3.842z M25.038,19.15 c0,5.049-8.678,8.912-9.047,9.072c-0.045,0.023-0.098,0.031-0.145,0.031c-0.051,0-0.098-0.008-0.146-0.031 c-0.365-0.16-9.046-4.023-9.046-9.072V6.811c0-0.199,0.161-0.363,0.362-0.363c7.229,0,8.482-2.756,8.494-2.783 c0.057-0.133,0.189-0.223,0.334-0.223c0.002,0,0.006,0,0.006,0c0.148,0.002,0.279,0.092,0.332,0.229 c0.012,0.025,1.24,2.777,8.494,2.777c0.201,0,0.361,0.164,0.361,0.363L25.038,19.15L25.038,19.15z"/></svg>',
        'user-x': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="8.5" cy="7" r="4"></circle><line x1="18" y1="8" x2="23" y2="13"></line><line x1="23" y1="8" x2="18" y2="13"></line></svg>',
        'check-circle': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>',
        'keyboard': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="16" x="2" y="4" rx="2"></rect><path d="M6 8h.001M10 8h.001M14 8h.001M18 8h.001M6 12h.001M10 12h.001M14 12h.001M18 12h.001M7 16h10"></path></svg>'
    };

    let cachedOverlay = null;
    let cachedElements = null;
    let activeResolver = null;

    function trapFocus(container, e) {
        const focusables = container.querySelectorAll(
            'a[href], button:not([disabled]), input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
        );
        if (focusables.length === 0) return;
        const first = focusables[0];
        const last = focusables[focusables.length - 1];
        if (e.shiftKey && document.activeElement === first) {
            e.preventDefault();
            last.focus();
        } else if (!e.shiftKey && document.activeElement === last) {
            e.preventDefault();
            first.focus();
        }
    }

    let previousActiveElement = null;
    let previousBodyOverflow = '';

    function getOrCreateDialog() {
        if (cachedOverlay && document.body.contains(cachedOverlay)) {
            return cachedElements;
        }

        const overlay = document.createElement('div');
        overlay.className = 'pro-confirm-overlay';
        overlay.style.display = 'none';
        overlay.innerHTML = `
            <div class="pro-confirm-card" role="dialog" aria-modal="true">
                <div class="pro-confirm-body">
                    <div class="pro-confirm-icon-badge" id="proConfirmIconBadge"></div>
                    <h3 class="pro-confirm-title" id="proConfirmTitle"></h3>
                    <p class="pro-confirm-desc" id="proConfirmDesc"></p>
                    <div class="pro-confirm-custom-content" id="proConfirmCustomContent"></div>
                </div>
                <div class="pro-confirm-footer">
                    <button type="button" class="pro-confirm-btn pro-confirm-btn-cancel" id="proConfirmCancelBtn"></button>
                    <button type="button" class="pro-confirm-btn pro-confirm-btn-confirm" id="proConfirmOkBtn"></button>
                </div>
            </div>
        `;

        document.body.appendChild(overlay);

        cachedOverlay = overlay;
        cachedElements = {
            overlay,
            card: overlay.querySelector('.pro-confirm-card'),
            badge: overlay.querySelector('#proConfirmIconBadge'),
            title: overlay.querySelector('#proConfirmTitle'),
            desc: overlay.querySelector('#proConfirmDesc'),
            customContent: overlay.querySelector('#proConfirmCustomContent'),
            footer: overlay.querySelector('.pro-confirm-footer'),
            cancelBtn: overlay.querySelector('#proConfirmCancelBtn'),
            okBtn: overlay.querySelector('#proConfirmOkBtn')
        };

        const cleanup = (result) => {

            promoteTemporarily(cachedElements.card, 'transform, opacity', 200);
            overlay.classList.remove('active');
            document.removeEventListener('keydown', onKey);
            document.body.style.overflow = previousBodyOverflow;
            if (previousActiveElement && typeof previousActiveElement.focus === 'function') {
                try { previousActiveElement.focus(); } catch (_) { }
            }
            setTimeout(() => {
                overlay.style.display = 'none';
                if (activeResolver) {
                    const resolve = activeResolver;
                    activeResolver = null;
                    resolve(result);
                }
            }, 180);
        };

        const onKey = (e) => {
            if (e.key === 'Tab') {
                trapFocus(cachedElements.card, e);
            } else if (e.key === 'Escape') {
                cleanup(false);
            } else if (e.key === 'Enter' && document.activeElement !== cachedElements.cancelBtn) {
                e.preventDefault();
                cleanup(true);
            }
        };

        cachedElements.cleanup = cleanup;
        cachedElements.onKey = onKey;

        cachedElements.cancelBtn.addEventListener('click', () => cleanup(false));
        cachedElements.okBtn.addEventListener('click', () => cleanup(true));
        overlay.addEventListener('click', (e) => {
            if (!cachedElements.card.contains(e.target)) cleanup(false);
        });

        return cachedElements;
    }

    function promoteTemporarily(el, props = 'transform, opacity', durationMs = 250) {
        if (!el || !el.style) return;
        el.style.willChange = props;
        setTimeout(() => {
            if (el && el.style) el.style.willChange = 'auto';
        }, durationMs);
    }

    function showConfirmDialog({
        title = 'آیا اطمینان دارید؟',
        description = '',
        htmlContent = '',
        confirmText = 'بله، تأیید',
        cancelText = 'خیر، انصراف',
        singleButton = false,
        variant = 'warning',
        iconType = 'shield-question'
    } = {}) {
        return new Promise((resolve) => {
            const els = getOrCreateDialog();
            activeResolver = resolve;

            previousActiveElement = document.activeElement;
            previousBodyOverflow = document.body.style.overflow;
            document.body.style.overflow = 'hidden';

            const iconSvg = icons[iconType] || icons['shield-question'];
            els.badge.className = `pro-confirm-icon-badge pro-confirm-badge-${variant}`;
            els.badge.innerHTML = iconSvg;

            els.title.textContent = title;

            if (description) {
                els.desc.textContent = description;
                els.desc.style.display = '';
            } else {
                els.desc.textContent = '';
                els.desc.style.display = 'none';
            }

            if (htmlContent) {
                els.customContent.innerHTML = htmlContent;
                els.customContent.style.display = '';
            } else {
                els.customContent.innerHTML = '';
                els.customContent.style.display = 'none';
            }

            if (singleButton || !cancelText) {
                els.cancelBtn.style.display = 'none';
                if (els.footer) els.footer.classList.add('single-btn');
            } else {
                els.cancelBtn.style.display = '';
                els.cancelBtn.textContent = cancelText;
                if (els.footer) els.footer.classList.remove('single-btn');
            }

            els.okBtn.textContent = confirmText;
            els.okBtn.className = `pro-confirm-btn pro-confirm-btn-confirm btn-var-${variant}`;

            promoteTemporarily(els.card, 'transform, opacity', 250);

            els.overlay.style.display = 'flex';
            void els.overlay.offsetWidth;
            els.overlay.classList.add('active');

            document.removeEventListener('keydown', els.onKey);
            document.addEventListener('keydown', els.onKey);

            setTimeout(() => {
                if (els.okBtn) els.okBtn.focus();
            }, 50);
        });
    }

    window.showConfirmDialog = showConfirmDialog;
})();
