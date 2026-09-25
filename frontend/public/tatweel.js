(function (global) {
    const RIGHT_ONLY = new Set(['ا', 'آ', 'أ', 'إ', 'د', 'ذ', 'ر', 'ز', 'ژ', 'و', 'ؤ']);
    const DUAL = new Set(['ب', 'پ', 'ت', 'ث', 'ج', 'چ', 'ح', 'خ', 'س', 'ش', 'ص', 'ض', 'ط', 'ظ', 'ع', 'غ', 'ف', 'ق', 'ک', 'گ', 'ل', 'م', 'ن', 'ه', 'ی', 'ي', 'ئ']);
    const TATWEEL = '\u0640';

    const CURATED = {
        'سامانه': 'سـامانه',
        'ستاد': 'سـتاد',
        'خبری': 'خـبری',
        'گزارش': 'گـزارش',
        'گزارشها': 'گـزارشها',
        'گزارشهای': 'گـزارش‌های',
        'گزارش‌ها': 'گـزارش‌ها',
        'گزارش‌های': 'گـزارش‌های',
        'مردمی': 'مـردمی',
        'کارتابل': 'کارتـابل',
        'محرمانه': 'محـرمانه',
        'حراست': 'حـراست',
        'داشبورد': 'داشـبورد',
        'آمار': 'آمـار',
        'تحلیل': 'تحـلیل',
        'داده‌ها': 'داده‌ها',
        'کاربران': 'کاربـران',
        'کاربر': 'کاربـر',
        'نشست‌های': 'نشـست‌های',
        'نشست‌ها': 'نشـست‌ها',
        'بازرسی': 'بازرسـی',
        'وضعیت': 'وضعیـت',
        'سلامت': 'سـلامت',
        'پشتیبانی': 'پشتیبـانی',
        'امنیت': 'امنیـت',
        'امنیتی': 'امنیـتی',
        'جزئیات': 'جزئیـات',
        'اطلاعات': 'اطـلاعات',
        'پیگیری': 'پیـگیری',
        'تنظیمات': 'تنـظیمات',
        'پرونده': 'پـرونده',
        'پرونده‌ها': 'پـرونده‌ها',
        'بررسی': 'بررسـی',
        'پاسخ': 'پـاسخ',
        'پاسخ‌ها': 'پـاسخ‌ها',
        'پاسخ‌های': 'پـاسخ‌های',
        'مفید': 'مفـید',
        'ارجاع': 'ارجـاع',
        'سوابق': 'سـوابق',
        'نمودار': 'نمـودار',
        'گردش': 'گـردش',
        'تغییر': 'تغـییر',
        'مدیریت': 'مدیـریت',
        'مدیریتی': 'مدیـریتی',
        'راهنما': 'راهنمـا',
        'سوالات': 'سـوالات',
        'متداول': 'متـداول',
        'گفتگو': 'گفتگـو',
        'تکمیلی': 'تکـمیلی',
        'طبقه‌بندی': 'طبقه‌بنـدی',
        'ثبت': 'ثـبت',
        'جدید': 'جـدید',
        'مراحل': 'مـراحل',
        'هویت': 'هویـت',
        'موضوع': 'مـوضوع',
        'مستندات': 'مستنـدات',
        'موفقیت': 'موفقیـت',
        'اعلانات': 'اعلانـات',
        'رویدادها': 'رویـدادها',
        'ارزیابی': 'ارزیـابی',
        'پیشرفته': 'پیشـرفته',
        'سیستم': 'سیسـتم',
        'پایگاه': 'پایگـاه',
        'ورود': 'ورود',
        'امن': 'امـن',
        'رسید': 'رسـید',
        'ارتباط': 'ارتبـاط',
        'مستقیم': 'مستقـیم'
    };

    const STOP_WORDS = new Set(['و', 'در', 'به', 'از', 'با', 'برای', 'تا', 'بر', 'یا', 'یک', 'این', 'آن', 'کل', 'همه', 'شد', 'است', 'شدن', 'شود', 'خود', 'ای', 'لاگ‌های', 'اعلان‌ها']);

    const config = {
        enabled: true,
        maxTatweelDesktop: 2,
        maxTatweelTablet: 1,
        maxTatweelMobile: 0,
        longHeadingCharLimit: 40,
        safetyThreshold: 0.92
    };

    function canConnect(a, b) {
        return DUAL.has(a) && (DUAL.has(b) || RIGHT_ONLY.has(b));
    }

    function enhanceWord(word, intensity) {
        if (!word || word.length < 2 || intensity <= 0) return word;
        const clean = word.trim();
        if (CURATED[clean]) {
            let res = CURATED[clean];
            if (intensity > 1) {
                res = res.replace(TATWEEL, TATWEEL.repeat(intensity));
            }
            return res;
        }
        for (let i = 0; i < clean.length - 1; i++) {
            const a = clean[i];
            const b = clean[i + 1];
            if (canConnect(a, b)) {
                return clean.slice(0, i + 1) + TATWEEL.repeat(intensity) + clean.slice(i + 1);
            }
        }
        return clean;
    }

    function getViewportIntensity(level) {
        const w = window.innerWidth;
        if (w <= 480) {
            return level === 'hero' ? 1 : 0;
        }
        if (w <= 768) {
            return level === 'hero' ? 1 : (level === 'section' ? 1 : 0);
        }
        if (level === 'hero') return 2;
        if (level === 'section') return 1;
        if (level === 'card') return 1;
        return 1;
    }

    function enhanceText(text, options) {
        options = options || {};
        const level = options.level || 'section';
        let intensity = typeof options.intensity === 'number' ? options.intensity : getViewportIntensity(level);
        if (!config.enabled || intensity <= 0 || !text) return text;
        const trimmed = text.trim();
        if (trimmed.length > config.longHeadingCharLimit) {
            intensity = 1;
        }
        const words = trimmed.split(/\s+/);
        if (words.length === 1) {
            return enhanceWord(words[0], intensity);
        }
        const maxWordsToElongate = words.length > 3 ? 2 : 1;
        let elongated = 0;
        const result = words.map(function (w) {
            if (STOP_WORDS.has(w) || elongated >= maxWordsToElongate) return w;
            if (CURATED[w]) {
                elongated++;
                return enhanceWord(w, intensity);
            }
            return w;
        });
        if (elongated === 0 && words.length > 0) {
            for (let i = 0; i < result.length; i++) {
                if (!STOP_WORDS.has(result[i]) && result[i].length >= 3) {
                    const candidate = enhanceWord(result[i], intensity);
                    if (candidate !== result[i]) {
                        result[i] = candidate;
                        break;
                    }
                }
            }
        }
        return result.join(' ');
    }

    function applyWrapperToNode(targetNode, rawText, level) {
        if (!rawText) return;
        targetNode.setAttribute('data-tatweel-raw', rawText);
        targetNode.setAttribute('data-tatweel-level', level);

        const existingWrapper = targetNode.querySelector('.persian-tatweel-wrapper');
        if (existingWrapper) {
            const visual = existingWrapper.querySelector('.tatweel-visual');
            const sr = existingWrapper.querySelector('.tatweel-sr');
            if (sr) sr.textContent = rawText;
            if (!config.enabled) {
                if (visual) visual.textContent = rawText;
                return;
            }
            const enhanced = enhanceText(rawText, { level: level });
            if (visual) visual.textContent = enhanced;
            if (targetNode.clientWidth > 0 && targetNode.scrollWidth > targetNode.clientWidth * config.safetyThreshold) {
                if (visual) visual.textContent = rawText;
            }
            return;
        }

        if (!config.enabled) {
            return;
        }

        const enhanced = enhanceText(rawText, { level: level });
        const wrapper = document.createElement('span');
        wrapper.className = 'persian-tatweel-wrapper';
        wrapper.setAttribute('data-tatweel-raw', rawText);

        const visual = document.createElement('span');
        visual.className = 'tatweel-visual';
        visual.setAttribute('aria-hidden', 'true');
        visual.textContent = enhanced;

        const sr = document.createElement('span');
        sr.className = 'tatweel-sr';
        sr.textContent = rawText;

        wrapper.appendChild(visual);
        wrapper.appendChild(sr);

        targetNode.textContent = '';
        targetNode.appendChild(wrapper);

        if (targetNode.clientWidth > 0 && targetNode.scrollWidth > targetNode.clientWidth * config.safetyThreshold) {
            visual.textContent = rawText;
        }
    }

    function enhanceElement(el, options) {
        if (!el || el.nodeType !== 1) return;
        if (el.getAttribute('data-tatweel') === 'false') return;
        if (el.closest('[data-tatweel="false"]') || el.closest('table, select, textarea, input, pre, code')) return;

        const level = (options && options.level) || el.getAttribute('data-tatweel-level') || (
            el.tagName === 'H1' ? 'hero' :
            el.tagName === 'H2' ? 'section' :
            el.tagName === 'H3' ? 'card' : 'subtle'
        );

        const innerSpan = el.querySelector('span:not(.persian-tatweel-wrapper):not(.badge):not(.notif-badge-pill):not(.superuser-badge-pill)');
        if (innerSpan && !innerSpan.querySelector('svg')) {
            let raw = innerSpan.getAttribute('data-tatweel-raw') || innerSpan.textContent.trim();
            if (raw) {
                applyWrapperToNode(innerSpan, raw, level);
                return;
            }
        }

        if (el.querySelector('svg')) {
            for (let i = 0; i < el.childNodes.length; i++) {
                const child = el.childNodes[i];
                if (child.nodeType === 3) {
                    const raw = child.nodeValue.trim();
                    if (raw.length > 1) {
                        const span = document.createElement('span');
                        span.textContent = raw;
                        el.replaceChild(span, child);
                        applyWrapperToNode(span, raw, level);
                        break;
                    }
                }
            }
            return;
        }

        let raw = el.getAttribute('data-tatweel-raw') || el.textContent.trim();
        if (raw) {
            applyWrapperToNode(el, raw, level);
        }
    }

    function init(container) {
        const root = container || document;
        const selector = '.persian-heading, .persian-heading--tatweel, h1, h2, h3, .modal-title, .kpi-label, .success-title, .step-label, .nexus-chat-title';
        const elements = root.querySelectorAll(selector);
        for (let i = 0; i < elements.length; i++) {
            enhanceElement(elements[i]);
        }
    }

    function recheckOverflow() {
        const elements = document.querySelectorAll('[data-tatweel-raw]');
        for (let i = 0; i < elements.length; i++) {
            const el = elements[i];
            const raw = el.getAttribute('data-tatweel-raw');
            const level = el.getAttribute('data-tatweel-level') || 'section';
            if (raw) {
                applyWrapperToNode(el, raw, level);
            }
        }
    }

    function setEnabled(flag) {
        config.enabled = Boolean(flag);
        recheckOverflow();
    }

    let resizeTimer = null;
    window.addEventListener('resize', function () {
        clearTimeout(resizeTimer);
        resizeTimer = setTimeout(recheckOverflow, 150);
    }, { passive: true });

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', function () {
            init();
        });
    } else {
        init();
    }

    global.TatweelTypography = {
        config: config,
        enhanceWord: enhanceWord,
        enhanceText: enhanceText,
        enhanceElement: enhanceElement,
        init: init,
        recheckOverflow: recheckOverflow,
        setEnabled: setEnabled
    };
})(window);
