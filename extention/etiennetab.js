var server = window.location.host == "localhost:8000" ? "http://localhost:8080" : "https://etiennetab.herokuapp.com";

fetch(server + '/gifs.json', {
    method: 'GET',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json'
    },
})
.then((resp) => resp.json())
.then((gifs) => {
    window.gifs = gifs;
    // pick a random gif
    let gif = gifs[Math.floor(Math.random() * gifs.length)]; 
    
    document.getElementById("src").setAttribute("src", gif); 
    let v = document.getElementById("v");
    v.addEventListener('loadstart', () => {
      v.play();
    });
    v.load();
})
.catch(function(err) {
    console.log(err);
});
