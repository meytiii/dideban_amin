(() => {
    let dismissed = false;

    const dismiss = () => {
        if (dismissed) return;
        dismissed = true;
        const p = document.getElementById('appPreloader');
        if (p) {
            p.classList.add('fade-out');
            setTimeout(() => {
                if (p.parentNode) p.parentNode.removeChild(p);
            }, 250);
        }
    };

    window.dismissAppPreloader = dismiss;

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', dismiss, { once: true });
    } else {
        dismiss();
    }

    setTimeout(dismiss, 2500);
})();
