import * as OS from "./obrolansubuh"

(function () {
	"use strict";

	window.addEventListener("WebComponentsReady", function (evt) {
		var mainDrawerPanel = document.querySelector("#main-drawer-panel"),
			postTitle = document.querySelector("input#post-title"),
		    mainTitle = document.querySelector("#main-panel-title");

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
			5000,
			true
		);

		ToastNotif.Show();

		elem.parentElement.removeChild(elem);
	});

	var publishButton = document.querySelector("#publish-post");
	if (publishButton) {
		publishButton.addEventListener("click", CreatePostSubmitListener(
			document.querySelector("input#post-title"),
			document.querySelector("#post-editor"),
			true
		));
	}

	var draftButton = document.querySelector("#save-draft");
	if (draftButton) {
		draftButton.addEventListener("click", CreatePostSubmitListener(
			document.querySelector("input#post-title"),
			document.querySelector("#post-editor"),
			false
		));
	}

	var updateButton = document.querySelector("#update-post");
	if (updateButton) {
		updateButton.addEventListener("click", CreatePutSubmitListener(
			document.querySelector("input#post-title"),
			document.querySelector("#post-editor"),
			document.querySelector("input#post-id"),
			document.querySelector("input#post-publish").value
		));
	}

	function CreatePutSubmitListener(titleElem, editorElem, idElem, publish) {
		return (evt) => {
			var id = idElem.value,
				data = {
					id        : parseInt(id),
					title     : titleElem.value,
					content   : editorElem.getEditorContent(),
					published : publish === "true"
				}, 
				parent = document.querySelector("#flash-container"),
				ToastNotif;

			$.ajax({
				url         : "/post/" + id + "/edit",
				type        : "PUT",
				data        : JSON.stringify(data),
				contentType : "application/json",
				success     : (data, textStatus, jqXHR) => {
					var text  = jqXHR.responseJSON["message"];

					ToastNotif = new OS.ToastNotification(parent, text, 5000, false);
				}
			}).always(() => { ToastNotif.Show(); });

			evt.preventDefault();
		}
	}

	function CreatePostSubmitListener(titleElem, editorElem, publish) {
		return (evt) => {
			var postData = {
				title   : titleElem.value,
				content : editorElem.getEditorContent(),
				publish : publish
			}, ToastNotif;

			$.post("/post/new", postData, (data, textStatus, jqXHR) => {
				window.location.replace(data["links"][0]["uri"]);
			})
			.fail((jqXHR, textStatus, errorThrown) => {
				var parent  = document.querySelector("#flash-container"),
				    message = jqXHR.responseJSON["messages"][0];

				ToastNotif = new OS.ToastNotification(parent, message, 5000, true);
			})
			.always(() => { ToastNotif.Show(); });

			evt.preventDefault();
		}
	}
})();