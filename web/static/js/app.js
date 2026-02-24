document.addEventListener('click', function(e) {
    if (e.target.classList.contains('suggest-item')) {
        const input = e.target.closest('.relative').querySelector('input');
        if (input) {
            input.value = e.target.dataset.value;
            e.target.parentElement.innerHTML = '';
            e.target.parentElement.classList.add('hidden');
        }
        return;
    }

        document.querySelectorAll('.suggest-dropdown').forEach(function(el) {
        el.classList.add('hidden');
    });
});

document.addEventListener('htmx:afterSwap', function(e) {
    if (e.detail.target.classList.contains('suggest-dropdown')) {
        if (e.detail.target.innerHTML.trim()) {
            e.detail.target.classList.remove('hidden');
        } else {
            e.detail.target.classList.add('hidden');
        }
    }
});

document.addEventListener('DOMContentLoaded', function() {
    var dateInputs = document.querySelectorAll('input[type="date"]');
    if (typeof flatpickr !== 'undefined') {
        dateInputs.forEach(function(input) {
            flatpickr(input, {
                dateFormat: 'Y-m-d',
                defaultDate: input.value || 'today'
            });
        });
    }
});

var style = document.createElement('style');
style.textContent = '@keyframes fadeIn { from { opacity: 0; background-color: #dcfce7; } to { opacity: 1; background-color: #f0fdf4; } }';
document.head.appendChild(style);
