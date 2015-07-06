(function () {
	window.addEventListener("WebComponentsReady", function (evt) {
		var mainDrawerPanel = document.querySelector("#main-drawer-panel"),
			postTitle = document.querySelector("input#post-title"),
		    mainTitle = document.querySelector("#PostTitlePanel");

		mainDrawerPanel.forceNarrow = true;
		mainDrawerPanel.closeDrawer();

		if (postTitle.value.length > 0) {
			mainTitle.textContent = postTitle.value;
			document.title = postTitle.value;
		}

		postTitle.addEventListener("keyup", function (evt) {
			var title = "Untitled";
			if (postTitle.value.length > 0) {
				title = postTitle.value;
			} 

			mainTitle.textContent = title;
			document.title = title;
		})
	});
})();