/**
 * Shared Modal Confirmation Dialog System (M-4)
 * Herasat Whistleblowing Portal - Education Department of Hamadan
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

    function showConfirmDialog({
        title = 'آیا اطمینان دارید؟',
        description = '',
        htmlContent = '',
        confirmText = 'بله، تأیید',
        cancelText = 'خیر، انصراف',
        variant = 'warning', // 'warning', 'danger', 'primary', 'success'
        iconType = 'shield-question' // 'logout', 'trash', 'shield-alert', 'shield-question', 'user-x', 'check-circle'
    } = {}) {
        return new Promise((resolve) => {
            document.querySelectorAll('.pro-confirm-overlay').forEach(el => el.remove());

            const iconSvg = icons[iconType] || icons['shield-question'];
            const overlay = document.createElement('div');
            overlay.className = 'pro-confirm-overlay';
            overlay.innerHTML = `
                <div class="pro-confirm-card" role="dialog" aria-modal="true">
                    <div class="pro-confirm-body">
                        <div class="pro-confirm-icon-badge pro-confirm-badge-${variant}">
                            ${iconSvg}
                        </div>
                        <h3 class="pro-confirm-title">${safeEscape(title)}</h3>
                        ${description ? `<p class="pro-confirm-desc">${safeEscape(description)}</p>` : ''}
                        ${htmlContent ? `<div class="pro-confirm-custom-content">${htmlContent}</div>` : ''}
                    </div>
                    <div class="pro-confirm-footer">
                        <button type="button" class="pro-confirm-btn pro-confirm-btn-cancel" id="proConfirmCancelBtn">
                            ${safeEscape(cancelText)}
                        </button>
                        <button type="button" class="pro-confirm-btn pro-confirm-btn-confirm btn-var-${variant}" id="proConfirmOkBtn">
                            ${safeEscape(confirmText)}
                        </button>
                    </div>
                </div>
            `;

            document.body.appendChild(overlay);

            // Reflow to ensure animation triggers
            void overlay.offsetWidth;
            overlay.classList.add('active');

            const card = overlay.querySelector('.pro-confirm-card');
            const cancelBtn = overlay.querySelector('#proConfirmCancelBtn');
            const okBtn = overlay.querySelector('#proConfirmOkBtn');

            function cleanup(result) {
                overlay.classList.remove('active');
                document.removeEventListener('keydown', onKey);
                setTimeout(() => {
                    overlay.remove();
                    resolve(result);
                }, 180);
            }

            function onKey(e) {
                if (e.key === 'Escape') cleanup(false);
            }

            cancelBtn.addEventListener('click', () => cleanup(false));
            okBtn.addEventListener('click', () => cleanup(true));
            overlay.addEventListener('click', (e) => {
                if (!card.contains(e.target)) cleanup(false);
            });
            document.addEventListener('keydown', onKey);

            setTimeout(() => okBtn.focus(), 50);
        });
    }

    window.showConfirmDialog = showConfirmDialog;
})();
