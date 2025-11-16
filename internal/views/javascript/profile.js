let canvas, ctx, img, scale = 1, offsetX = 0, offsetY = 0;
let isDragging = false, startX, startY;
let originalFile;

function initAvatarEditor(event) {
    const file = event.target.files[0];
    if (!file) return;

    originalFile = file;
    const reader = new FileReader();

    reader.onload = function (e) {
        document.getElementById('avatar-display').style.display = 'none';
        document.getElementById('avatar-editor').style.display = 'block';
        document.getElementById('upload-btn').style.display = 'inline-block';
        document.getElementById('cancel-btn').style.display = 'inline-block';

        canvas = document.getElementById('avatar-canvas');
        ctx = canvas.getContext('2d');
        img = new Image();

        img.onload = function () {
            const imgAspect = img.width / img.height;
            if (imgAspect > 1) {
                scale = 300 / img.height;
                offsetX = (300 - img.width * scale) / 2;
                offsetY = 0;
            } else {
                scale = 300 / img.width;
                offsetX = 0;
                offsetY = (300 - img.height * scale) / 2;
            }
            drawImage();
            setupControls();
        };

        img.src = e.target.result;
    };

    reader.readAsDataURL(file);
}

function drawImage() {
    ctx.clearRect(0, 0, 300, 300);
    ctx.drawImage(img, offsetX, offsetY, img.width * scale, img.height * scale);
}

function setupControls() {
    const slider = document.getElementById('scale-slider');
    const scaleValue = document.getElementById('scale-value');

    slider.oninput = function () {
        const newScale = this.value / 100;
        const centerX = 150;
        const centerY = 150;
        offsetX = centerX - (centerX - offsetX) * (newScale / scale);
        offsetY = centerY - (centerY - offsetY) * (newScale / scale);
        scale = newScale;

        scaleValue.textContent = this.value + '%';
        drawImage();
    };

    canvas.onmousedown = function (e) {
        isDragging = true;
        startX = e.offsetX;
        startY = e.offsetY;
    };

    canvas.onmousemove = function (e) {
        if (isDragging) {
            const dx = e.offsetX - startX;
            const dy = e.offsetY - startY;
            offsetX += dx;
            offsetY += dy;
            startX = e.offsetX;
            startY = e.offsetY;
            drawImage();
        }
    };

    canvas.onmouseup = function () {
        isDragging = false;
    };

    canvas.onmouseleave = function () {
        isDragging = false;
    };
    canvas.ontouchstart = function (e) {
        e.preventDefault();
        isDragging = true;
        const touch = e.touches[0];
        const rect = canvas.getBoundingClientRect();
        startX = touch.clientX - rect.left;
        startY = touch.clientY - rect.top;
    };

    canvas.ontouchmove = function (e) {
        e.preventDefault();
        if (isDragging) {
            const touch = e.touches[0];
            const rect = canvas.getBoundingClientRect();
            const x = touch.clientX - rect.left;
            const y = touch.clientY - rect.top;
            const dx = x - startX;
            const dy = y - startY;
            offsetX += dx;
            offsetY += dy;
            startX = x;
            startY = y;
            drawImage();
        }
    };

    canvas.ontouchend = function () {
        isDragging = false;
    };
}

function uploadAvatar() {
    const finalCanvas = document.createElement('canvas');
    finalCanvas.width = 200;
    finalCanvas.height = 200;
    const finalCtx = finalCanvas.getContext('2d');
    const sourceX = 0;
    const sourceY = 0;
    const sourceSize = 300;

    finalCtx.drawImage(canvas, sourceX, sourceY, sourceSize, sourceSize, 0, 0, 200, 200);
    finalCanvas.toBlob(function (blob) {
        const formData = new FormData();
        formData.append('avatar', blob, 'avatar.jpg');
        
        const csrfToken = document.querySelector('input[name="gorilla.csrf.Token"]').value;
        formData.append('gorilla.csrf.Token', csrfToken);

        fetch('/profile/upload-avatar', {
            method: 'POST',
            body: formData
        })
            .then(response => {
                if (response.redirected) {
                    window.location.href = response.url;
                } else {
                    return response.text();
                }
            })
            .catch(error => {
                alert('Error download: ' + error);
            });
    }, 'image/jpeg', 0.9);
}

function cancelEdit() {
    document.getElementById('avatar-display').style.display = 'block';
    document.getElementById('avatar-editor').style.display = 'none';
    document.getElementById('upload-btn').style.display = 'none';
    document.getElementById('cancel-btn').style.display = 'none';
    document.getElementById('avatar-input').value = '';
}
