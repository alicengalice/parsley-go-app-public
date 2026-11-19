const toast = document.getElementById('toast');
const primaryButton = document.querySelector('.primary-button');

if (primaryButton) {
  primaryButton.addEventListener('click', () => {
    toast.classList.add('show');
    window.clearTimeout(window.toastTimer);
    window.toastTimer = window.setTimeout(() => {
      toast.classList.remove('show');
    }, 1800);
  });
}
