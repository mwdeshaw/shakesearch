const PAGE_KEY = "currentPage";
const ROWS_KEY = "currentRows";

const Controller = {
  getState: (key) => {
    return localStorage.getItem(key);
  },

  setState: (key, val) => {
    localStorage.setItem(key, val);
  },

  search: (ev, page) => {
    ev.preventDefault();
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}&p=${page}`).then(
      (response) => {
        response.json().then((results) => {
          Controller.updateTable(results);
          Controller.setState(PAGE_KEY, page);
        });
      }
    );
  },

  initialSearch: (ev) => {
    Controller.setState(ROWS_KEY, JSON.stringify([]));
    Controller.search(ev, 0);
  },

  loadMore: (ev) => {
    const currentPage = Controller.getState(PAGE_KEY);
    Controller.search(ev, parseInt(currentPage) + 1);
  },

  updateTable: (results) => {
    const table = document.getElementById("table-body");
    const prevRows = JSON.parse(Controller.getState(ROWS_KEY)) || [];
    const rows = [...prevRows];
    for (let result of results) {
      rows.push(`<tr><td>${result}</td></tr>`);
    }
    Controller.setState(ROWS_KEY, JSON.stringify(rows));
    table.innerHTML = rows;
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.initialSearch);
document
  .getElementById("load-more")
  .addEventListener("click", Controller.loadMore);
