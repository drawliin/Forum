// Get buttons element for both posts and comments
let btnPost = document.querySelectorAll(".react-post")
let btnComment= document.querySelectorAll(".react-comment")

//Add event listener on post buttons
btnPost.forEach(button => {
    button.addEventListener("click", function (e) {
        e.preventDefault();

        //prevent redirection to post/id fron card's onclick
        e.stopPropagation();
        
        const postID = this.dataset.postId;
        const value = this.dataset.value;

        sendPostReaction(postID, value,this);
    });
});

//Add event listener on Comment buttons
btnComment.forEach(button => {
    button.addEventListener("click", function (e) {
        e.preventDefault();

        const commentID = this.dataset.commentId;
        const value = this.dataset.value;
        
        sendCommentReaction(commentID, value, this)
    });
});

//Sends a post request when triggered with the value of the button
//it then receives the Json with the number of likes/dislikes
//and display them on the button using the id attribute
function sendPostReaction(postID, value, button) {
    
    button.disabled = true;

    fetch(`/post/${postID}/react`, {
        method: "POST",
        credentials: "include",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded",
        },
        body: `value=${value}`
    })
        .then(res => {
            if (!res.ok) {
                return res.text().then(html => {
                document.open();
                document.writeln(html);
                document.close();})  
            }
                // throw new Error("Request failed");

            const contentType = res.headers.get("content-type");

            if (!contentType|| !contentType.includes("application/json")) {
                window.location.href = "/login";
                return;
            }
            return res.json();
        })
        .then(data => {
            if (!data) return;
            
            document.getElementById(`likes-${postID}`).textContent = data.likes;
            document.getElementById(`dislikes-${postID}`).textContent = data.dislikes;
        })
        .catch(error => console.error("Error:", error))
        .finally(() => {
            button.disabled = false;
        });
}

//Sends a post request when triggered with the value of the button
//using the comment id 
//it then receives the Json with the number of likes/dislikes
//and display them on the button using the id attribute
function sendCommentReaction(commentID, value, button) {
    
    button.disabled = true;

    fetch(`/comment/${commentID}/react`, {
        method: "POST",
        credentials: "include",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded",
        },
        body: `value=${value}`
    })
        .then(res => {
            if (!res.ok) {
                return res.text().then(html => {
                document.open();
                document.writeln(html);
                document.close();})  
            }
            const contentType = res.headers.get("content-type");

            if (!contentType|| !contentType.includes("application/json")) {
                window.location.href = "/login";
                return;
            }
            return res.json();
        })
        .then(data => {
            if (!data) return;
            document.getElementById(`likes-${commentID}-comment`).textContent = data.likes;
            document.getElementById(`dislikes-${commentID}-comment`).textContent = data.dislikes;
        })
        .catch(error => console.error("Error:", error))
        .finally(() => {
            button.disabled = false;
        });
}
