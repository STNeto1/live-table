class Controls extends HTMLElement {
	constructor() {
		super();

		document.addEventListener("DOMContentLoaded", () => {
			this.update();

			htmx.on("htmx:wsAfterMessage", () => {
				this.update();
			});
		});
	}

	update() {
		// create page buttons
		const table = document.getElementById("records");
		const pagesDiv = document.getElementById("pages");
		pagesDiv.innerHTML = "";

		if (!table) {
			alert("No table found");
			return;
		}

		const currentPage = parseInt(table.dataset.currentPage);
		const totalPages = parseInt(table.dataset.pages);

		const $pages = this.getPagingRange(currentPage, {
			total: totalPages,
			length: 7,
		});

		for (const page of $pages) {
			const form = this.createPageForm(page, currentPage);
			pagesDiv.appendChild(form);

			htmx.process(form);
		}

		// create page size buttons
		const pagesSizesDiv = document.getElementById("page_sizes");
		pagesSizesDiv.innerHTML = "";

		const pageSize = parseInt(table.dataset.pageSize);
		const $options = [10, 20, 50, 100];

		for (const opt of $options) {
			const form = this.createPageSizeForm(opt, pageSize);
			pagesSizesDiv.appendChild(form);

			htmx.process(form);
		}
	}

	createPageForm(page, currentPage) {
		const form = document.createElement("form");
		form.id = `page-form-${page}`;
		form.setAttribute("ws-send", "");
		form.innerHTML = `
            <input type="hidden" name="event" value="change_page"/>
            <input type="hidden" name="to_page" value="${page}"/>
            <button type="submit" ${
							page === currentPage ? "disabled" : ""
						}>${page}</button>
        `;
		return form;
	}

	createPageSizeForm(opt, currentSize) {
		const form = document.createElement("form");
		form.id = `page-size-form-${opt}`;
		form.setAttribute("ws-send", "");
		form.innerHTML = `
            <input type="hidden" name="event" value="change_page_size"/>
            <input type="hidden" name="page_size" value="${opt}"/>
            <button type="submit" ${
							opt === currentSize ? "disabled" : ""
						}> ${opt} </button>

        `;
		return form;
	}

	getPagingRange(current, { min = 1, total = 20, length = 5 } = {}) {
		if (length > total) length = total;

		let start = current - Math.floor(length / 2);
		start = Math.max(start, min);
		start = Math.min(start, min + total - length);

		return Array.from({ length: length }, (_, i) => start + i);
	}
}

customElements.define("wc-controls", Controls);
