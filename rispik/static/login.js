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
    //console.log("fillSelect called", selectID)
    let select = document.getElementById(selectID);
    fetch(baseURLWithProtocol + relativeURL)
        .then(response => response.json())
        .then(data => {
            select.innerHTML = "";
            let option = document.createElement("option");
            option.innerText = title;
            select.options.add(option);
            for (let i = 0; i < data.length; i++) {
                const val = data[i];
                let option = document.createElement("option");
                option.innerText = val;
                select.options.add(option);
                if (i === 0 && !document.getElementById(selectID).selectedIndex)
                    select.options.selectedIndex = i + 1;
            }
            //console.log("fillSelect completed", selectID, select.options);
            let urlParams = new URLSearchParams(window.location.search);
            if (urlParams.has(selectID)) {
                setSelectedOption(selectID, urlParams.get(selectID));
            } else if (localStorage.getItem("project")) {
                setSelectedOption(selectID, localStorage.getItem(selectID));
            }
            enableLogin();
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
    //console.log("setSelectedOption", selectID, value);
    let select = document.getElementById(selectID);
    let options = select.children;
    let foundValue = false;
    for (let i = 0; i < options.length; i++) {
        if (options[i].value === value) {
            select.selectedIndex = i;
            foundValue = true;
        }
    }
    if (!foundValue) {
        let option = document.createElement("option");
        option.innerText = value;
        option.checked = true;
        select.options.add(option);
        select.selectedIndex = options.length-1;
    }
    //console.log("setSelectedOption new index", document.getElementById(selectID).selectedIndex);
}

window.onload = function () {
    let selects = document.getElementsByTagName("select");
    for (let i = 0; i < selects.length; i++) {
        let select = selects[i];
        select.addEventListener("change", enableLogin);
    }

    fillSelect("/list/users", "Användare", "user");
    fillSelect("/list/projects", "Projekt", "project");
    fillSelect("/list/sessions", "Session", "session");
}
