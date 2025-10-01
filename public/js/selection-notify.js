   window.addEventListener('DOMContentLoaded', function () {
      document.querySelector("form").addEventListener("submit", function (e) {
        const sel1 = document.getElementById("selection1").value.trim();
        const sel2 = document.getElementById("selection2").value.trim();

        if (sel1 === "" && sel2 === "") {
          e.preventDefault(); // stop form submission
          alert("Please fill at least one of the selections.");
        }
      });

      document.getElementById("selection1").addEventListener("change", function () {
        const selection2 = document.getElementById("selection2");
        selection2.style.display = "";
        selection2.classList.add("active");
      });
      document.getElementById('submitBtn').addEventListener('click', function (e) {
        var topic = document.getElementById('topic').value;
        var body = document.getElementById('body').value;
        var sel1 = document.getElementById('selection1').value;
        // Prevent only whitespace for both fields
        if (!topic.trim()) {
          alert('Thread topic cannot be empty or only spaces.');
          e.preventDefault();
          return;
        }
        if (!body.trim()) {
          alert('Thread body cannot be empty or only spaces.');
          e.preventDefault();
          return;
        }
        if (!sel1) {
          alert('Please select a category.');
          e.preventDefault();
          return;
        }
        // Form will submit if validation passes
      });
    });