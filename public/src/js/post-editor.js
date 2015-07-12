import * as OS from "./obrolansubuh"

(function () {
	"use strict";

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

		var ToastNotif = new OS.ToastNotification(
			document.querySelector("#flash-container"),
			error["error"],
			5000
		);

		ToastNotif.Show();

		elem.parentElement.removeChild(elem);
	});

	var publishButton = document.querySelector("#publish-post");
	publishButton.addEventListener("click", function (evt) {
		var postTitle  = document.querySelector("input#post-title"),
			postEditor = document.querySelector("#post-editor"),
			postData   = {
				title   : postTitle.value,
				content : postEditor.getEditorContent(),
				publish : true
			};

		var ToastNotif; // for Toast Notification in case something goes wrong.
		$.post("/post/new", postData, function (data, textStatus, jqXHR) {
			window.location.replace(data["links"][0]["uri"]);
		})
		.fail(function (jqXHR, textStatus, errorThrown) {
			var parent = document.querySelector("#flash-container");

			ToastNotif = new OS.ToastNotification(parent, jqXHR.responseText, 5000);
		})
		.always(function () {
			ToastNotif.Show();
		});

		evt.preventDefault();
	});
})();