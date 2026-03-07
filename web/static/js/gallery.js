function enableEdit(id) {
    const display = document.getElementById(`title-display-${id}`);
    const input = document.getElementById(`title-input-${id}`);
    display.style.display = 'none';
    input.style.display = 'block';
    input.focus();
}

function saveTitle(id) {
    const display = document.getElementById(`title-display-${id}`);
    const input = document.getElementById(`title-input-${id}`);
    const newTitle = input.value.trim();

    if (!newTitle || newTitle === display.innerText) {
        input.style.display = 'none';
        display.style.display = 'block';
        return;
    }

    display.innerText = newTitle;
    input.style.display = 'none';
    display.style.display = 'block';

    fetch(`/edit/${id}?title=${encodeURIComponent(newTitle)}`, { method: 'POST' });
}

function filterVideos() {
    const filter = document.getElementById('searchInput').value.toLowerCase();
    const rows = document.querySelectorAll('#video-table tbody tr');
    rows.forEach(row => {
        const title = row.querySelector('.video-title-text').innerText.toLowerCase();
        row.style.display = title.includes(filter) ? '' : 'none';
    });
}

function toggleView(viewType) {
    const wrapper = document.getElementById('wrapper');
    const listBtn = document.getElementById('listBtn');
    const gridBtn = document.getElementById('gridBtn');

    if (viewType === 'grid') {
        wrapper.classList.add('grid-view');
        gridBtn.classList.add('active');
        listBtn.classList.remove('active');
    } else {
        wrapper.classList.remove('grid-view');
        listBtn.classList.add('active');
        gridBtn.classList.remove('active');
    }
}

// --- NEW HEART POP LOGIC ---
function toggleLike(btn, id) {
    btn.classList.toggle('liked');
    const svg = btn.querySelector('svg');
    svg.classList.add('heart-pop');

    setTimeout(() => {
        svg.classList.remove('heart-pop');
    }, 300);

    // Placeholder for database save logic
    console.log(`Video ${id} toggle like.`);
}