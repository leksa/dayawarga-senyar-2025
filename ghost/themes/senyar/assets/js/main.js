/**
 * Senyar Theme - Main JavaScript
 * Siklon Senyar Blog
 */

(function() {
    'use strict';

    // Mobile navigation toggle
    function initMobileNav() {
        const header = document.querySelector('.site-header');
        const nav = document.querySelector('.site-nav');

        if (!header || !nav) return;

        // Create hamburger button
        const hamburger = document.createElement('button');
        hamburger.className = 'nav-toggle';
        hamburger.innerHTML = '<span></span><span></span><span></span>';
        hamburger.setAttribute('aria-label', 'Toggle Navigation');

        // Only show on mobile
        hamburger.style.display = 'none';

        hamburger.addEventListener('click', function() {
            nav.classList.toggle('is-open');
            hamburger.classList.toggle('is-active');
        });

        // Insert before nav
        header.querySelector('.container').insertBefore(hamburger, nav);

        // Handle responsive
        function checkWidth() {
            if (window.innerWidth <= 768) {
                hamburger.style.display = 'block';
            } else {
                hamburger.style.display = 'none';
                nav.classList.remove('is-open');
                hamburger.classList.remove('is-active');
            }
        }

        window.addEventListener('resize', checkWidth);
        checkWidth();
    }

    // Lazy load images
    function initLazyLoad() {
        if ('IntersectionObserver' in window) {
            const images = document.querySelectorAll('img[data-src]');

            const imageObserver = new IntersectionObserver(function(entries, observer) {
                entries.forEach(function(entry) {
                    if (entry.isIntersecting) {
                        const img = entry.target;
                        img.src = img.dataset.src;
                        img.removeAttribute('data-src');
                        imageObserver.unobserve(img);
                    }
                });
            });

            images.forEach(function(img) {
                imageObserver.observe(img);
            });
        }
    }

    // Smooth scroll for anchor links
    function initSmoothScroll() {
        document.querySelectorAll('a[href^="#"]').forEach(function(anchor) {
            anchor.addEventListener('click', function(e) {
                const target = document.querySelector(this.getAttribute('href'));
                if (target) {
                    e.preventDefault();
                    target.scrollIntoView({
                        behavior: 'smooth',
                        block: 'start'
                    });
                }
            });
        });
    }

    // Reading progress indicator
    function initReadingProgress() {
        const postContent = document.querySelector('.post-content');
        if (!postContent) return;

        const progressBar = document.createElement('div');
        progressBar.className = 'reading-progress';
        progressBar.innerHTML = '<div class="reading-progress-bar"></div>';
        document.body.appendChild(progressBar);

        const bar = progressBar.querySelector('.reading-progress-bar');

        window.addEventListener('scroll', function() {
            const contentTop = postContent.offsetTop;
            const contentHeight = postContent.offsetHeight;
            const scrollTop = window.scrollY;
            const windowHeight = window.innerHeight;

            const progress = Math.min(
                Math.max((scrollTop - contentTop + windowHeight) / contentHeight, 0),
                1
            );

            bar.style.width = (progress * 100) + '%';
        });
    }

    // Initialize
    document.addEventListener('DOMContentLoaded', function() {
        initMobileNav();
        initLazyLoad();
        initSmoothScroll();
        initReadingProgress();
    });
})();
