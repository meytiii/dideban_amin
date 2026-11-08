(() => {
    let dismissed = false;

    const applyVideoPlaybackRate = () => {
        ['heroMotionVideo', 'successMotionVideo'].forEach(id => {
            const v = document.getElementById(id);
            if (v) {
                v.defaultPlaybackRate = 1.25;
                v.playbackRate = 1.25;
            }
        });
    };

    const dismiss = () => {
        if (dismissed) return;
        dismissed = true;
        applyVideoPlaybackRate();
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
        document.addEventListener('DOMContentLoaded', () => {
            applyVideoPlaybackRate();
            dismiss();
        }, { once: true });
    } else {
        applyVideoPlaybackRate();
        dismiss();
    }

    setTimeout(dismiss, 2500);
})();