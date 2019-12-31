function bunker_logout() {
  localStorage.removeItem("xtoken");
  localStorage.removeItem("xtoken");
  localStorage.removeItem("login");
  document.location = "/";
}

function dateFormat(value, row, index) {
  //return moment(value).format('DD/MM/YYYY');
  var d = new Date(parseInt(value) * 1000);
  let f_date =
    d.getFullYear() +
    "-" +
    ("0" + (d.getMonth() + 1)).slice(-2) +
    "-" +
    ("0" + d.getDate()).slice(-2) +
    " " +
    ("0" + d.getHours()).slice(-2) +
    ":" +
    ("0" + d.getMinutes()).slice(-2) +
    ":" +
    ("0" + d.getSeconds()).slice(-2);
  //return d.toUTCString();
  return f_date;
}

function shortString(value, row, index) {
  if (!value) {
    return "N/A";
  }
  return "<span title='" + value + "' class='shortstring'>" + value + "</span>";
}