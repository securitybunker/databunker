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

function prepareDropdownMenu(menu) {
  const currentFile = document.location.pathname.split(/[\/]+/).pop();
  var code = '';
  for (var index = 0; index < menu.length; index++) {
    const name = menu[index]["name"];
    const file = menu[index]["file"];
    const style = (file == currentFile) ? ' active' : '';
    code += '<a class="dropdown-item'+style+'" href="'+file+'">'+name+'</a>'+"\n";
  }
  return code;
}

function prepareMenu(menu) {
  const currentFile = document.location.pathname.split(/[\/]+/).pop();
  var code = '';
  for (var index = 0; index < menu.length; index++) {
    const name = menu[index]["name"];
    if (menu[index]["dropdown"]) {
      code += '<li class="nav-item dropdown">'+
        '<a class="nav-link dropdown-toggle" href="#" id="dropdown-'+name+'" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">'+name+'</a>'+
        '<div class="dropdown-menu" aria-labelledby="dropdown-'+name+'">'+
        prepareDropdownMenu( menu[index]["dropdown"]) + '</div></li>';
    } else {
      const file = menu[index]["file"];
      const style = (file == currentFile) ? ' active' : '';
      code += '<li class="nav-item">'+
         '<a class="nav-item nav-link'+style+'" href="'+file+'">'+name+'</a></li>'+"\n";
    }
  }
  return code;
}

function showAdminMenu() {
  const code = prepareMenu(adminMenu);
  const m = document.getElementById("admin-menu");
  if (m) {
    m.innerHTML = code;
  }

}

function showUserMenu() {
  const code = prepareMenu(userMenu);
  document.write(code);
}

function getUserStartPage() {
  return userMenu[0].file;
}

function showSuccess(msg) {
  const st = document.getElementById("status-message");
  if (st) {
    st.innerHTML = `<div class="alert alert-primary" id="success-alert">`+
      `<button type="button" class="close" data-dismiss="alert">x</button>`+
      `<strong>Success! </strong> `+msg+`</div>`;
    $("#success-alert").fadeTo(2000, 500).fadeOut(500);
  }
}

function showError(msg) {
  const st = document.getElementById("status-message");
  if (st) {
    st.innerHTML = `<div class="alert alert-warning" id="error-alert">`+
      `<button type="button" class="close" data-dismiss="alert">x</button>`+
      `<strong>Error! </strong> `+msg+`</div>`;
    $("#success-alert").fadeTo(2000, 500).fadeOut(500);
  }
}

function loadAgreements(method, address, cb) {
  var xhr1 = new XMLHttpRequest();
  xhr1.open('GET', "/v1/agreements/" + method + "/" + address);
  xhr1.setRequestHeader("X-Bunker-Token", xtoken)
  xhr1.setRequestHeader('Content-type', 'application/json');
  xhr1.onload = function () {
    if (xhr1.status === 200) {
      var data = JSON.parse(xhr1.responseText);
      if (cb) {
        cb(data);
      } else {
        console.log("loadAgreements cb is empty")
      }
    } else if (xhr1.status > 400 && xhr1.status < 404) {
      document.location = "/";
    }
  }
  xhr1.send();
}

function acceptAgreement(method, address, brief, options, cb) {
  var xhr1 = new XMLHttpRequest();
  var params = '';
  if (options) {
    params = JSON.stringify(options);
  }
  xhr1.open('POST', "/v1/agreement/" + brief + "/" + method + "/" + address);
  xhr1.setRequestHeader("X-Bunker-Token", xtoken)
  xhr1.setRequestHeader('Content-type', 'application/json');
  xhr1.onload = function() {
    if (xhr1.status === 200) {
      var data = JSON.parse(xhr1.responseText);
      if (cb) {
        cb(data);
      }
      return
    } else if (xhr1.status === 401) {
      document.location = "/";
    }
  }
  xhr1.send(params);
}
