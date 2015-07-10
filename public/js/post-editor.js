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
		var error = evt.detail.jqXHR.responseJSON["files"][0],
			// current uploading image
		    elem  = document.querySelector(".medium-insert-active");

		alert(error["error"]);

		elem.parentElement.removeChild(elem);
	});

	var publishButton = document.querySelector("#publish-post");
	publishButton.addEventListener("click", function (evt) {
		var postTitle  = document.querySelector("input#post-title"),
			postEditor = document.querySelector("#post-editor"),
			postData   = {
				title   : postTitle.value,
				content : postEditor.getEditorContent()
			};

		$.post("/post/new", postData, function (data, textStatus, jqXHR) {
			window.location.replace(data["links"][0]["uri"]);
		})
		.fail(function (jqXHR, textStatus, errorThrown) {
			var parent = document.querySelector("#flash-container");

			var notif = document.createElement("paper-toast"),
			    label = document.createElement("span");

			notif.setAttribute("duration", "5000");
			notif.classList.add("error");

			label.setAttribute("id", "label");
			label.classList.add("style-scope");
			label.classList.add("paper-toast");

			label.textContent = jqXHR.responseText;

			notif.appendChild(label);
			parent.appendChild(notif);
		})
		.always(function () {
			// timeout so the notification can appear more smoothly
			setTimeout(function () {
				document.querySelector("paper-toast").show();
			}, 1000);
		});

		evt.preventDefault();
	});
})();