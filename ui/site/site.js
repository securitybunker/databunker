function bunker_logout() {
  localStorage.removeItem("xtoken");
  localStorage.removeItem("login");
  document.location = "/";
}

var ui_configuration;
function loadUIConfiguration() {
  if (ui_configuration) {
    return ui_configuration;
  }
  var xhr10 = new XMLHttpRequest();
  xhr10.open('GET', "/v1/sys/uiconfiguration", false);
  xhr10.onload = function () {
    if (xhr10.status === 200) {
      console.log(xhr10.responseText);
      var data = JSON.parse(xhr10.responseText);
      if (data.status == "ok") {
        ui_configuration = data.ui;
      }
    }
  }
  xhr10.send();
  return ui_configuration;
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