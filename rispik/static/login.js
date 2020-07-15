"use strict";

const baseURL = window.location.host;
const baseURLWithProtocol = window.location.protocol + '//' + baseURL;// + window.location.pathname.replace(/\/$/g,"");

var respeaker = document.getElementById("role_respeaker");
var admin = document.getElementById("role_admin");
var respeakerSettings = document.getElementById("respeaker_settings");

respeaker.addEventListener("change", toggleSettingsPerRole);
admin.addEventListener("change", toggleSettingsPerRole);
respeaker.addEventListener("change", enableLogin);
admin.addEventListener("change", enableLogin);

function toggleSettingsPerRole() {
    if (respeaker.checked)
        respeakerSettings.classList.remove("hidden");
    else
        respeakerSettings.classList.add("hidden");
}

function fillSelect(relativeURL, title, selectID) {
    let users = fetch(baseURLWithProtocol + relativeURL)
        .then(response => response.json())
        .then(data => {
            document.getElementById(selectID).innerHTML = "";
            let option = document.createElement("option");
            option.innerText = title;
            document.getElementById(selectID).appendChild(option);
            for (let i=0;i<data.length;i++) {
                const val = data[i];
                let option = document.createElement("option");
                option.innerText = val;
                if (i === 0 && !document.getElementById(selectID).selectedIndex)
                    option.selected = true;
                document.getElementById(selectID).appendChild(option);
            }
        })
        .catch(error => {
            console.log("Couldn't list " + selectID + "s", error);
        });

}

document.getElementById("login").addEventListener("click", function (evt) {

    if (evt.target.getAttribute("disabled")) return;

    if (respeaker.checked) {
        let u = document.getElementById("user")[document.getElementById("user").selectedIndex];
        let p = document.getElementById("project")[document.getElementById("project").selectedIndex];
        let s = document.getElementById("session")[document.getElementById("session").selectedIndex];

        localStorage.setItem("user", u.value);
        localStorage.setItem("project", p.value);
        localStorage.setItem("session", s.value);

        window.location.replace(baseURLWithProtocol + "/main.html");
    }
    else {
        alert("Admin login is not implemented");
    }
});

function enableLogin(evt) {
    let enable = false
    if (admin.checked)
        enable = true;
    else {
        let u = document.getElementById("user");
        let p = document.getElementById("project");
        let s = document.getElementById("session");
        if (u.selectedIndex > 0 && p.selectedIndex > 0 && s.selectedIndex > 0)
            enable = true;
        else
            enable = false;
    }
    if (enable) {
        document.getElementById("login").classList.remove("disabled");
        document.getElementById("login").removeAttribute("disabled");
    } else {
        document.getElementById("login").classList.add("disabled");
        document.getElementById("login").setAttribute("disabled", "disabled");
    }
}

function setSelectedOption(selectID, value) {
    console.log("setSelectedOption", selectID, value);
    let options = document.getElementById(selectID).getElementsByTagName("option");
    let foundValue = false;
    for (let i = 0; i < options.length; i++) {
        if (options[i].value === value)
            options[i].checked = true;
            foundValue = true;
    }
    if (!foundValue) {
        let option = document.createElement("option");
        option.innerText = value;
        option.checked = true;
        document.getElementById(selectID).appendChild(option);
    }
}

window.onload = function () {
    fillSelect("/list/users", "Användare", "user");
    fillSelect("/list/projects", "Projekt", "project");
    fillSelect("/list/sessions", "Session", "session");

    let selects = document.getElementsByTagName("select");
    for (let i = 0; i < selects.length; i++) {
        let select = selects[i];
        select.addEventListener("change", enableLogin);
    }

    let urlParams = new URLSearchParams(window.location.search);
    if (localStorage.getItem("project")) {
        setSelectedOption("project", localStorage.getItem("project"));
    }
    if (urlParams.has('project')) {
        setSelectedOption("project", urlParams.get("project"));
    }
    if (localStorage.getItem("session")) {
        setSelectedOption("session", localStorage.getItem("session"));
    }
    if (urlParams.has('session')) {
        setSelectedOption("session", urlParams.get("session"));
    }
    if (localStorage.getItem("user")) {
        setSelectedOption("user", localStorage.getItem("user"));
    }
    if (urlParams.has('user')) {
        setSelectedOption("user", urlParams.get("user"));
    }

    enableLogin();

}
