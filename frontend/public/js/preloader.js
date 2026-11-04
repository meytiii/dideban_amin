(() => {
    let dismissed = false;

    const dismiss = () => {
        if (dismissed) return;
        dismissed = true;
        const p = document.getElementById('appPreloader');
        if (p && !p.classList.contains('fade-out')) {
            p.classList.add('fade-out');
            setTimeout(() => {
                if (p.parentNode) p.parentNode.removeChild(p);
            }, 450);
        }
    };

    window.dismissAppPreloader = dismiss;

    const waitForCriticalImages = () => {
        const imgs = Array.from(document.querySelectorAll('img[loading="eager"], .admin-crest-img, .workflow-banner-img'));
        const pending = imgs.filter(img => !img.complete || img.naturalWidth === 0);
        if (pending.length === 0) return Promise.resolve();
        return Promise.all(pending.map(img => new Promise(res => {
            const finish = () => {
                img.removeEventListener('load', finish);
                img.removeEventListener('error', finish);
                res();
            };
            img.addEventListener('load', finish, { once: true });
            img.addEventListener('error', finish, { once: true });
        })));
    };

    const waitForFonts = () => {
        if (!document.fonts || !document.fonts.ready) return Promise.resolve();
        return document.fonts.ready.catch(() => {});
    };

    const waitForLoad = () => {
        if (document.readyState === 'complete') return Promise.resolve();
        return new Promise(res => {
            window.addEventListener('load', res, { once: true });
        });
    };

    const checkReady = () => {
        Promise.all([waitForLoad(), waitForFonts(), waitForCriticalImages()]).then(() => {
            if (typeof window.__publicStartupGate === 'function') {
                window.__publicStartupGate().then(dismiss).catch(dismiss);
            } else if (window.__publicStartupPromise) {
                window.__publicStartupPromise.then(dismiss).catch(dismiss);
            } else {
                dismiss();
            }
        }).catch(dismiss);
    };

    if (document.readyState === 'complete') {
        checkReady();
    } else {
        window.addEventListener('load', checkReady, { once: true });
        document.addEventListener('DOMContentLoaded', checkReady, { once: true });
    }

    setTimeout(dismiss, 7000);
})();
