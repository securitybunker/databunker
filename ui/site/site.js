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
      //console.log(xhr10.responseText);
      var data = JSON.parse(xhr10.responseText);
      if (data && data.status == "ok") {
        ui_configuration = data.ui;
      }
    }
  }
  xhr10.send();
  return ui_configuration;
}

function displayFooterLinks()
{
  conf = loadUIConfiguration();
  if (conf["TermOfServiceTitle"]) {
    document.write("<div class='text-center'><a href='"+conf["TermOfServiceLink"]+"'>"+conf["TermOfServiceTitle"]+"</a></div>" );
  }
  if (conf["TermOfServiceTitle"]) {
    document.write("<div class='text-center'><a href='"+conf["PrivacyPolicyLink"]+"'>"+conf["PrivacyPolicyTitle"]+"</a></div>" );
  }
  if (conf["CompanyTitle"]) {
    document.write("<div class='text-center'><a href='"+conf["CompanyLink"]+"'>"+conf["CompanyTitle"]+"</a></div>" );
  }
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

function showAlert(message) {
	var alertModal =
    $('<div class="modal fade" role="dialog"><div class="modal-dialog" role="document"><div class="modal-content">' +
      '<div class="modal-header">' +
      '<h5 class="modal-title">Alert</h5>' +
      '<button type="button" class="close" data-dismiss="modal" aria-label="Close">' +
      '<span aria-hidden="true">&times;</span></button>' +
      '</div>' +
      '<div class="modal-body">' +
         '<div class="alert alert-warning" role="alert">' + message + '</div>' +
      '</div>' +
      '<div class="modal-footer">' +
        '<a href="#" class="btn" data-dismiss="modal">OK</a>' +
      '</div>' +
      '</div></div></div>');

    alertModal.modal('show');
}

function showForm(title, message, btn) {
	if (!btn) {
		btn = "";
	}
	var formModal =
    $('<div class="modal fade" role="dialog"><div class="modal-dialog modal-lg" role="document"><div class="modal-content">' +
      '<div class="modal-header">' +
      '<h5 class="modal-title">' + title + '</h5>' +
      '<button type="button" class="close" data-dismiss="modal" aria-label="Close">' +
      '<span aria-hidden="true">&times;</span></button>' +
      '</div>' +
      '<div class="modal-body">' +
         message +
      '</div>' +
      '<div class="modal-footer">' + btn +
        '<a href="#" class="btn" data-dismiss="modal">Cancel</a>' +
      '</div>' +
      '</div></div></div>');
    formModal.modal('show');
	return formModal;
}
