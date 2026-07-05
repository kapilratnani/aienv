// Mobile Navigation
const navToggle = document.querySelector('.nav-toggle');
const navLinks = document.querySelector('.nav-links');

if (navToggle && navLinks) {
  navToggle.addEventListener('click', () => {
    navLinks.classList.toggle('active');
    navToggle.setAttribute('aria-expanded', navLinks.classList.contains('active'));
  });
}

// Copy to Clipboard
function copyToClipboard(text, button) {
  if (!navigator.clipboard) {
    fallbackCopyTextToClipboard(text, button);
    return;
  }
  
  navigator.clipboard.writeText(text).then(() => {
    showCopyFeedback(button);
  }).catch(err => {
    console.error('Failed to copy:', err);
  });
}

function fallbackCopyTextToClipboard(text, button) {
  const textArea = document.createElement('textarea');
  textArea.value = text;
  textArea.style.position = 'fixed';
  textArea.style.opacity = '0';
  document.body.appendChild(textArea);
  textArea.select();
  try {
    document.execCommand('copy');
    showCopyFeedback(button);
  } catch (err) {
    console.error('Fallback: Oops, unable to copy', err);
  }
  document.body.removeChild(textArea);
}

function showCopyFeedback(button) {
  const originalText = button.textContent;
  button.textContent = 'Copied!';
  button.style.color = 'var(--accent-green)';
  
  setTimeout(() => {
    button.textContent = originalText;
    button.style.color = '';
  }, 2000);
}

// Setup copy buttons for code blocks
document.querySelectorAll('.copy-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    const container = btn.closest('.code-block, .install-command');
    if (container) {
      const pre = container.querySelector('pre');
      const code = container.querySelector('code');
      let text = '';
      if (pre) {
        text = pre.textContent;
      } else if (code) {
        text = code.textContent;
      }
      if (text) copyToClipboard(text, btn);
    }
  });
});

// Recipe Accordions
document.querySelectorAll('.recipe-header').forEach(header => {
  header.addEventListener('click', () => {
    const content = header.nextElementSibling;
    const toggle = header.querySelector('.recipe-toggle');
    if (content) {
      content.classList.toggle('active');
    }
    if (toggle) {
      toggle.classList.toggle('active');
    }
  });
});

// Scroll Animations
const observerOptions = {
  root: null,
  rootMargin: '0px',
  threshold: 0.1
};

const observer = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      entry.target.classList.add('visible');
    }
  });
}, observerOptions);

document.querySelectorAll('.fade-in').forEach(el => {
  observer.observe(el);
});

// Smooth scroll for nav links
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
  anchor.addEventListener('click', function(e) {
    e.preventDefault();
    const targetId = this.getAttribute('href').substring(1);
    const target = document.getElementById(targetId);
    if (target) {
      target.scrollIntoView({
        behavior: 'smooth',
        block: 'start'
      });
    }
  });
});

// Close mobile menu on link click
if (navLinks) {
  document.querySelectorAll('.nav-links a').forEach(link => {
    link.addEventListener('click', () => {
      navLinks.classList.remove('active');
      if (navToggle) {
        navToggle.setAttribute('aria-expanded', 'false');
      }
    });
  });
}