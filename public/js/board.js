var markers = [];

mapboxgl.accessToken = 'pk.eyJ1IjoibWF0dGE5MDAxIiwiYSI6ImNsZWcweHk1ajAyM3Y0M3BobTJ5anQyMjEifQ.fxWRoGxTSyz1T1vw0I3R8w';
const map = new mapboxgl.Map({
    container: 'map', // container ID
    style: 'mapbox://styles/mapbox/streets-v12', // style URL
    center: [-87.623177, 41.881832],
    zoom: 9 // starting zoom
});

const cardList = document.getElementById("card-list")

function createMarker(markerData) {

    var popUp = new mapboxgl.Popup({ offset: 25 }).setHTML("<h3>" + markerData.name + "</h3><p>" + markerData.description + "</p>");

    // Create marker object
    var marker = new mapboxgl.Marker({
        color: "#0D6EFD",
    }).setLngLat([markerData.lng, markerData.lat])
        .setPopup(popUp)
        .addTo(map);

    // Create Bootstrap card with javascript directly
    const card = document.createElement("div");
    card.classList.add("card");
    card.classList.add("p-4");
    card.classList.add("m-4");

    const cardBody = document.createElement("div");
    cardBody.classList.add("card-body");

    const cardTitle = document.createElement("h5");
    cardTitle.classList.add("card-title");
    cardTitle.textContent = markerData.name;
    cardBody.append(cardTitle);

    const cardText = document.createElement("p");
    cardText.classList.add("card-text");
    cardText.textContent = markerData.description;
    cardBody.append(cardText);

    card.appendChild(cardBody);

    card.id = "list-card-" + markerData.id;
    card.tabIndex = markerData.id;

    card.onclick = function () {
        map.flyTo({ center: marker.getLngLat(), zoom: 15 });
        scrollToCard(markerData);
    }

    cardList.appendChild(card);

    marker.getElement().addEventListener('click', function () {
        map.flyTo({ center: marker.getLngLat(), zoom: 15 });
        scrollToCard(markerData);
    });
}

function scrollToCard(markerData) {
    const myCollapsibleDiv = document.getElementById("collapseFilters");
    myCollapsibleDiv.classList.add("show");

    const cardFocus = document.getElementById("list-card-" + markerData.id);

    // Calculate the center position of the screen
    const centerX = window.innerWidth / 2;
    const centerY = window.innerHeight / 2;

    // Calculate the position of the div relative to the screen
    const rect = cardFocus.getBoundingClientRect();
    const divX = rect.x + rect.width / 2;
    const divY = rect.y + rect.height / 2;

    // Calculate the scroll offset needed to center the div
    const scrollX = divX - centerX;
    const scrollY = divY - centerY;

    // Apply the scroll offset to the window
    window.scrollBy({ left: scrollX, top: scrollY, behavior: "smooth" });

    highlightCard(markerData);
}

function highlightCard(markerData) {
    var items = cardList.querySelectorAll('div');

    items.forEach(function (item) {
        item.classList.remove('shadow');
        item.classList.remove('text-bg-primary');
    });

    items.forEach(function (item) {
        if (item.id == 'list-card-' + markerData.id) {
            item.classList.add('shadow');
            item.classList.add('text-bg-primary');
        }
    });
}

markers.forEach(function (markerData) {
    createMarker(markerData);
});

