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
    let a = gifs[Math.floor(Math.random() * gifs.length)]; 
    let handle = a[0];
    let statusID = a[1]
    let gif = a[2];
    
    document.getElementById("src").setAttribute("src", gif); 
    let v = document.getElementById("v");
    v.addEventListener('loadstart', () => {
      v.play();
    });
    v.load();

    let h = document.getElementById("h");
    h.setAttribute("href", `https://twitter.com/${handle}/status/${statusID}`);
    h.innerHTML = `@${handle}`;

})
.catch(function(err) {
    console.log(err);
});


let timeout = null;
document.addEventListener("DOMContentLoaded", function(){
    console.log("what");
    document.getElementById("v").addEventListener('mousemove', function() {
	    console.log("r");
	    document.getElementById("h").classList.add('show');
	    clearTimeout(timeout);

	    timeout = setTimeout(function() {
		document.getElementById("h").classList.remove('show');
	    }, 2000);
    });
});

