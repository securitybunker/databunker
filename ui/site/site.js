function bunker_logout()
{
    localStorage.removeItem("xtoken");
    localStorage.removeItem("xtoken");
    localStorage.removeItem("login");
    document.location = "/";
}

function dateFormat(value, row, index) {
    //return moment(value).format('DD/MM/YYYY');
    var d = new Date(parseInt(value) * 1000);
    let f_date = d.getFullYear() + "-" + (d.getMonth() + 1) + "-" + d.getDate() +
      " " + d.getHours() + ":" + d.getMinutes() + ":" + d.getSeconds()
    //return d.toUTCString();
    return f_date;
}