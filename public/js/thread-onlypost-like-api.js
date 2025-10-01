// Post voting functions
function likePost(postId) {
  console.log("Like post button clicked for post:", postId);
  const likeBtn = document.querySelector(
    `button[onclick="likePost(${postId})"]`
  );
  const dislikeBtn = document.querySelector(
    `button[onclick="dislikePost(${postId})"]`
  );

  if (likeBtn) {
    likeBtn.disabled = true;
  }

  if (dislikeBtn) {
    dislikeBtn.disabled = true;
  }

  fetch("/api/post/" + postId + "/like", {
    method: "POST",
    credentials: "same-origin", // Include cookies
  })
    .then((response) => {
      console.log("Like post response status:", response.status);
      if (response.status === 401) {
        alert("Please log in to like posts.");
        throw new Error("Unauthorized - Please log in");
      }
      if (!response.ok) {
        throw new Error("Network response was not ok");
      }
      return response.json();
    })
    .then((data) => {
      console.log("Like post response data:", data);
      if (data) {
        updatePostVoteDisplay(postId, data);
      }
    })
    .catch((error) => {
      console.log("=== POST LIKE ERROR CAUGHT ===");
      console.log("Error object:", error);
      if (error.message.includes("Unauthorized")) {
        console.log("User is not authorized - login required");
      } else if (error.message.includes("401")) {
        alert("Please log in to like posts.");
      } else if (error.message.includes("404")) {
        alert("Post not found.");
      } else {
        console.log("Generic error caught in like post function:");
        alert(
          "Error processing like. Please make sure you are logged in and try again."
        );
      }
    })
    .finally(() => {
      // Re-enable buttons
      if (likeBtn) {
        likeBtn.disabled = false;
      }
      if (dislikeBtn) {
        dislikeBtn.disabled = false;
      }
    });
}

function dislikePost(postId) {
  console.log("Dislike post button clicked for post:", postId);
  const likeBtn = document.querySelector(
    `button[onclick="likePost(${postId})"]`
  );
  const dislikeBtn = document.querySelector(
    `button[onclick="dislikePost(${postId})"]`
  );

  if (likeBtn) {
    likeBtn.disabled = true;
  }
  if (dislikeBtn) {
    dislikeBtn.disabled = true;
  }

  fetch("/api/post/" + postId + "/dislike", {
    method: "POST",
    credentials: "same-origin", // Include cookies
  })
    .then((response) => {
      console.log("Dislike post response status:", response.status);
      if (response.status === 401) {
        alert("Please log in to dislike posts.");
        throw new Error("Unauthorized - Please log in");
      }
      if (!response.ok) {
        throw new Error("Network response was not ok");
      }
      return response.json();
    })
    .then((data) => {
      console.log("Dislike post response data:", data);
      if (data) {
        updatePostVoteDisplay(postId, data);
      }
    })
    .catch((error) => {
      console.log("=== POST DISLIKE ERROR CAUGHT ===");
      console.log("Error object:", error);
      if (error.message.includes("Unauthorized")) {
        console.log("User is not authorized - login required");
      } else if (error.message.includes("401")) {
        alert("Please log in to dislike posts.");
      } else if (error.message.includes("404")) {
        alert("Post not found.");
      } else {
        console.log("Generic error caught in dislike post function:");
        alert(
          "Error processing dislike. Please make sure you are logged in and try again."
        );
      }
    })
    .finally(() => {
      // Re-enable buttons
      if (likeBtn) {
        likeBtn.disabled = false;
      }
      if (dislikeBtn) {
        dislikeBtn.disabled = false;
      }
    });
}

// Update post vote display based on server response
function updatePostVoteDisplay(postId, data) {
  console.log(
    "updatePostVoteDisplay called for post:",
    postId,
    "with data:",
    data
  );

  // Update like/dislike counts
  const likesSpan = document.getElementById(`post-likes-${postId}`);
  const dislikesSpan = document.getElementById(`post-dislikes-${postId}`);

  if (likesSpan) {
    likesSpan.textContent = data.likes || 0;
  }
  if (dislikesSpan) {
    dislikesSpan.textContent = data.dislikes || 0;
  }

  // Try multiple ways to find the buttons
  let likeBtn = document.querySelector(`button[onclick="likePost(${postId})"]`);
  let dislikeBtn = document.querySelector(
    `button[onclick="dislikePost(${postId})"]`
  );

  // Fallback: try using data-post-id attribute
  if (!likeBtn) {
    likeBtn = document.querySelector(
      `button[data-post-id="${postId}"].like-btn`
    );
  }
  if (!dislikeBtn) {
    dislikeBtn = document.querySelector(
      `button[data-post-id="${postId}"].dislike-btn`
    );
  }

  // Update button styles based on user vote status
  if (likeBtn) {
    if (data.userLiked) {
      likeBtn.classList.remove("btn-success");
      likeBtn.classList.add("btn-success", "active"); // Blue when user liked
    } else {
      likeBtn.classList.remove("active");
      likeBtn.classList.add("btn-success"); // Normal green state
    }
  }

  if (dislikeBtn) {
    if (data.userDisliked) {
      dislikeBtn.classList.remove("btn-danger");
      dislikeBtn.classList.add("btn-danger", "active"); // Blue when user disliked
    } else {
      dislikeBtn.classList.remove("active");
      dislikeBtn.classList.add("btn-danger"); // Normal red state
    }
  }
}

// Load initial post vote counts when page loads
document.addEventListener("DOMContentLoaded", function () {
  // Get all post IDs from the page
  const postButtons = document.querySelectorAll("button[data-post-id]");
  postButtons.forEach((button) => {
    const postId = button.getAttribute("data-post-id");
    if (postId) {
      // Fetch initial vote status for each post
      fetch("/api/post/" + postId + "/status", {
        method: "GET",
        credentials: "same-origin",
      })
        .then((response) => response.json())
        .then((data) => {
          updatePostVoteDisplay(postId, data);
        })
        .catch((error) => {
          console.log(
            "Error fetching post vote status for post",
            postId,
            ":",
            error
          );
        });
    }
  });

  // Get all thread IDs from the page
  const threadButtons = document.querySelectorAll("button[data-thread-id]");
  threadButtons.forEach((button) => {
    const threadId = button.getAttribute("data-thread-id");
    if (threadId) {
      // Fetch initial vote status for each thread
      fetch("/api/thread/" + threadId + "/status", {
        method: "GET",
        credentials: "same-origin",
      })
        .then((response) => response.json())
        .then((data) => {
          updateThreadVoteDisplay(threadId, data);
        })
        .catch((error) => {
          console.log(
            "Error fetching thread vote status for thread",
            threadId,
            ":",
            error
          );
        });
    }
  });
});

// Thread voting functions
function likeThread(threadId) {
  console.log("Like thread button clicked for thread:", threadId);
  const likeBtn = document.querySelector(
    `button[onclick="likeThread(${threadId})"]`
  );
  const dislikeBtn = document.querySelector(
    `button[onclick="dislikeThread(${threadId})"]`
  );

  if (likeBtn) {
    likeBtn.disabled = true;
  }
  if (dislikeBtn) {
    dislikeBtn.disabled = true;
  }

  fetch("/api/thread/" + threadId + "/like", {
    method: "POST",
    credentials: "same-origin", // Include cookies
  })
    .then((response) => {
      console.log("Like thread response status:", response.status);
      if (response.status === 401) {
        alert("Please log in to like threads.");
        throw new Error("Unauthorized - Please log in");
      }
      if (!response.ok) {
        throw new Error("Network response was not ok");
      }
      return response.json();
    })
    .then((data) => {
      console.log("Like thread response data:", data);
      if (data) {
        updateThreadVoteDisplay(threadId, data);
      }
    })
    .catch((error) => {
      console.log("=== THREAD LIKE ERROR CAUGHT ===");
      console.log("Error object:", error);
      if (error.message.includes("Unauthorized")) {
        console.log("User is not authorized - login required");
      } else if (error.message.includes("401")) {
        alert("Please log in to like threads.");
      } else if (error.message.includes("404")) {
        alert("Thread not found.");
      } else {
        console.log("Generic error caught in like thread function:");
        alert(
          "Error processing like. Please make sure you are logged in and try again."
        );
      }
    })
    .finally(() => {
      // Re-enable buttons
      if (likeBtn) {
        likeBtn.disabled = false;
      }
      if (dislikeBtn) {
        dislikeBtn.disabled = false;
      }
    });
}

function dislikeThread(threadId) {
  console.log("Dislike thread button clicked for thread:", threadId);
  const likeBtn = document.querySelector(
    `button[onclick="likeThread(${threadId})"]`
  );
  const dislikeBtn = document.querySelector(
    `button[onclick="dislikeThread(${threadId})"]`
  );

  if (likeBtn) {
    likeBtn.disabled = true;
  }
  if (dislikeBtn) {
    dislikeBtn.disabled = true;
  }

  fetch("/api/thread/" + threadId + "/dislike", {
    method: "POST",
    credentials: "same-origin", // Include cookies
  })
    .then((response) => {
      console.log("Dislike thread response status:", response.status);
      if (response.status === 401) {
        alert("Please log in to dislike threads.");
        throw new Error("Unauthorized - Please log in");
      }
      if (!response.ok) {
        throw new Error("Network response was not ok");
      }
      return response.json();
    })
    .then((data) => {
      console.log("Dislike thread response data:", data);
      if (data) {
        updateThreadVoteDisplay(threadId, data);
      }
    })
    .catch((error) => {
      console.log("=== THREAD DISLIKE ERROR CAUGHT ===");
      console.log("Error object:", error);
      if (error.message.includes("Unauthorized")) {
        console.log("User is not authorized - login required");
      } else if (error.message.includes("401")) {
        alert("Please log in to dislike threads.");
      } else if (error.message.includes("404")) {
        alert("Thread not found.");
      } else {
        console.log("Generic error caught in dislike thread function:");
        alert(
          "Error processing dislike. Please make sure you are logged in and try again."
        );
      }
    })
    .finally(() => {
      // Re-enable buttons
      if (likeBtn) {
        likeBtn.disabled = false;
      }
      if (dislikeBtn) {
        dislikeBtn.disabled = false;
      }
    });
}

// Update thread vote display based on server response
function updateThreadVoteDisplay(threadId, data) {
  console.log(
    "updateThreadVoteDisplay called for thread:",
    threadId,
    "with data:",
    data
  );

  // Update like/dislike counts
  const likesSpan = document.getElementById(`thread-likes-${threadId}`);
  const dislikesSpan = document.getElementById(`thread-dislikes-${threadId}`);

  if (likesSpan) {
    likesSpan.textContent = data.likes || 0;
  }
  if (dislikesSpan) {
    dislikesSpan.textContent = data.dislikes || 0;
  }

  // Try multiple ways to find the buttons
  let likeBtn = document.querySelector(
    `button[onclick="likeThread(${threadId})"]`
  );
  let dislikeBtn = document.querySelector(
    `button[onclick="dislikeThread(${threadId})"]`
  );

  // Fallback: try using data-thread-id attribute
  if (!likeBtn) {
    likeBtn = document.querySelector(
      `button[data-thread-id="${threadId}"].thread-like-btn`
    );
  }
  if (!dislikeBtn) {
    dislikeBtn = document.querySelector(
      `button[data-thread-id="${threadId}"].thread-dislike-btn`
    );
  }

  // Update button styles based on user vote status
  if (likeBtn) {
    if (data.userLiked) {
      likeBtn.classList.remove("btn-success");
      likeBtn.classList.add("btn-info"); // Highlighted when user liked
    } else {
      likeBtn.classList.remove("btn-info");
      likeBtn.classList.add("btn-success"); // Normal state
    }
  }

  if (dislikeBtn) {
    if (data.userDisliked) {
      dislikeBtn.classList.remove("btn-danger");
      dislikeBtn.classList.add("btn-warning"); // Highlighted when user disliked
    } else {
      dislikeBtn.classList.remove("btn-warning");
      dislikeBtn.classList.add("btn-danger"); // Normal state
    }
  }
}
