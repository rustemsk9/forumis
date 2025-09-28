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
