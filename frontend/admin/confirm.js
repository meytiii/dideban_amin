/**
 * Shared Modal Confirmation Dialog System (M-4) [PERF-FIX-17] [PERF-FIX-23]
 * Herasat Whistleblowing Portal - Education Department of Hamadan
 * Lazy DOM creation with element caching to eliminate GC thrash and HTML parsing on each call.
 */

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
        'trash': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>',
        'shield-alert': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>',
        'shield-question': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path><path d="M12 8a2 2 0 0 1 2 2c0 1.2-1.5 1.8-1.5 2.5"></path><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>',
        'user-x': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="8.5" cy="7" r="4"></circle><line x1="18" y1="8" x2="23" y2="13"></line><line x1="23" y1="8" x2="18" y2="13"></line></svg>',
        'check-circle': '<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>'
    };

    let cachedOverlay = null;
    let cachedElements = null;
    let activeResolver = null;

    // [HARDEN-8] Trap focus inside modal dialog
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
            cancelBtn: overlay.querySelector('#proConfirmCancelBtn'),
            okBtn: overlay.querySelector('#proConfirmOkBtn')
        };

        const cleanup = (result) => {
            // [PERF-FIX-28] Temporarily promote modal card layer during exit transition
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

    // [PERF-FIX-28] Dynamic will-change promotion helper with auto-cleanup
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

            els.cancelBtn.textContent = cancelText;
            els.okBtn.textContent = confirmText;
            els.okBtn.className = `pro-confirm-btn pro-confirm-btn-confirm btn-var-${variant}`;

            // [PERF-FIX-28] Temporarily promote modal card layer during open transition
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
