console.log("test");

let btn = document.querySelectorAll(".react")
btn.forEach(button => {
    button.addEventListener("click", function (e) {
        // e.preventDefault();

        const postID = this.dataset.postId;
        const value = this.dataset.value;

        sendReaction(postID, value);
    });
});

function sendReaction(postID, value) {
    fetch(`/post/${postID}/react`, {
        method: "POST",
        credentials:"include",
        headers: {
            "Content-Type":"application/x-www-form-urlencoded",
        },
        body:`value=${value}`
    })
        .then(res => {
            if (!res.ok) {
                console.error("Request failed");
                return;
            }
            res.json()
        })
        .then(data => {
            console.log(data);
            
            document.getElementById(`likes-${postID}`).textContent = data.likes;
            document.getElementById(`dislikes-${postID}`).textContent = data.dislikes;
        })
        // .then(response => {
        //     if (response.ok) {
        //         console.log("reaction updated")
        //         location.reload();
        //     } else {
        //         console.error("Failed to react")
        // }
        // })
        .catch(error=> console.error("Error:",error))
}
