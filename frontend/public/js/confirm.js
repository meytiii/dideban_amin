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
        'shield-alert': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>',
        'shield-question': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path><path d="M12 8a2 2 0 0 1 2 2c0 1.2-1.5 1.8-1.5 2.5"></path><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>',
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
