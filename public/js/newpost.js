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

	var postEditor = document.querySelector("#post-editor");
	postEditor.addEventListener("image-upload-failed", function (evt) {
		// This comes from the plugin we use. Event data is exposed 
		// naked to the user (us) so we can have better control.
		// See https://github.com/blueimp/jQuery-File-Upload/wiki/Options
		var error = evt.detail.jqXHR.responseJSON["files"][0];
		alert(error["error"]);
	});
})();