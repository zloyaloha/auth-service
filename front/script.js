document.addEventListener('DOMContentLoaded', () => {
    const showLoginBtn = document.getElementById('show-login-btn');
    const showRegisterBtn = document.getElementById('show-register-btn');
    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');

    showLoginBtn.addEventListener('click', () => {
        loginForm.classList.remove('hidden');
        registerForm.classList.add('hidden');

        showLoginBtn.classList.add('active');
        showRegisterBtn.classList.remove('active');
    });

    showRegisterBtn.addEventListener('click', () => {
        loginForm.classList.add('hidden');
        registerForm.classList.remove('hidden');

        showLoginBtn.classList.remove('active');
        showRegisterBtn.classList.add('active');
    });

    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const firstName = registerForm.elements['register-first-name'].value;
        const lastName = registerForm.elements['register-last-name'].value;
        const email = registerForm.elements['register-email'].value;
        const password = registerForm.elements['register-password'].value;
        const confirmPassword = registerForm.elements['register-confirm-password'].value;

        if (password !== confirmPassword) {
            alert('Пароли не совпадают!');
            return;
        }

        try {
            const response = await fetch('http://localhost:8080/v1/registrate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ firstName, lastName, email, password })
            });

            const result = await response.json();
            if (response.ok) {
                alert(result.message);
            } else {
                alert('Ошибка' + result.message)
            }
        } catch (err) {
            alert('Сетевая ошибка' + err.message)
        }
    });

    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const email = loginForm.elements['login-email'].value
        const password = loginForm.elements['login-password'].value

        try {
            const response = await fetch('http://localhost:8080/v1/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({email, password, 'app_id': 1})
            });

            const result = await response.json();
            if (response.ok) {
                alert("Успешно " + result.message)
            } else {
                alert('Ошибка ' + result.message)
            }
        } catch (err) {
            alert('Network error' + err.message)
        }
    });
});

