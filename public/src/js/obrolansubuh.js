export class ToastNotification {
	constructor(parentElement, content, duration, isError) {
    	this.notif = document.createElement("paper-toast"),
    	this.label = document.createElement("span");

    	this.notif.setAttribute("duration", duration);

        if (isError) {
        	this.notif.classList.add("error");
        }

    	this.label.setAttribute("id", "label");
    	this.label.classList.add("style-scope");
    	this.label.classList.add("paper-toast");

    	this.label.textContent = content;

    	this.notif.appendChild(this.label);
    	parentElement.appendChild(this.notif);

    	return this;
	}
	Show(element) {
		setTimeout(() => { this.notif.show(); }, 500);
	}
}