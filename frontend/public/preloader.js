(() => {
    const dismiss = () => {
        const p = document.getElementById('appPreloader');
        if (p && !p.classList.contains('fade-out')) {
            p.classList.add('fade-out');
            setTimeout(() => {
                if (p.parentNode) p.parentNode.removeChild(p);
            }, 350);
        }
    };
    if (document.readyState === 'interactive' || document.readyState === 'complete') {
        requestAnimationFrame(dismiss);
    } else {
        document.addEventListener('DOMContentLoaded', dismiss, { once: true });
        window.addEventListener('load', dismiss, { once: true });
        setTimeout(dismiss, 500);
    }
})();
