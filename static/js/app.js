function stopAjax(jqXHRWBD, jqXHRWT) {
    jqXHRWBD.abort();
    jqXHRWT.abort();
    document.betform.submit();
}

function popup(type, title, text) {
    Swal.fire({
        icon: type,
        title: title,
        text: text,
    })
}

function temppopup(type, title, text, pos, sec) {
    Swal.fire({
        icon: type,
        title: title,
        text: text,
        showConfirmButton: false,
        timer: sec * 1000,
        position: pos,
    })
}

function mobileWaitingPhase(jqXHRWT, jqXHRWBD) {
    jqXHRWT = $.ajax({
        url: '/ajax/mobilewaitingphase',
        dataType: 'json',
        success: async function (data) {
            console.log("mobileWaitingPhase")
            switch (data["func"]) {
                case "prefrop":
                    // Prefrop
                    notie.alert({ type: 'info', text: data["text"], stay: true })
                    break;
                case "frop":
                    // Frop
                    notie.alert({ type: 'info', text: data["text"], stay: true })
                    frop(data["cards"]);
                    break;
                case "turn":
                    // Turn
                    notie.alert({ type: 'info', text: data["text"], stay: true })
                    turn(data["card"]);
                    break;
                case "river":
                    // River
                    notie.alert({ type: 'info', text: data["text"], stay: true })
                    river(data["card"]);
                    break;
                case "result":
                    // Result
                    notie.alert({ type: 'info', text: data["text"], stay: true })
                    // Pop Up
                    Swal.fire({
                        icon: "info",
                        title: "Info",
                        text: "Show Down",
                        confirmButtonText: '<a href=' + data["URL"] + ' style="text-decoration:none; color:white; font-size:large; font-weight: bold;">Show Down</a>'
                    });
                    break;
                case "reset":
                    // Reset
                    console.log(jqXHRWBD)
                    stopAjax(jqXHRWBD, jqXHRWT)
                    document.location.href = data["redirect"];
                    break;
                case "popup":
                    // Pop Up
                    Swal.fire({
                        icon: "error",
                        title: "Something went wrong!",
                        text: "Please click OK",
                        confirmButtonText: '<a href="/remotepoker" style="text-decoration:none; color:white; font-size:large; font-weight: bold;">OK</a>'
                    });
                    break;
            }
            mobileWaitingPhase();
        },
    });
    return jqXHRWT
}

function remoteWaitingPhase(jqXHRWT, jqXHRWBD) {
    jqXHRWT = $.ajax({
        url: '/ajax/mobilewaitingphase',
        data: { from: "remotepoker" },
        dataType: 'json',
        success: async function (data) {
            console.log("remoteWaitingPhase")
            $('#playerBetSize').text(data["bet"])
            $('#playerStack').text(data["stack"])
            switch (data["func"]) {
                case "prefrop":
                    // Prefrop
                    notie.alert({ type: 'info', text: data["text"], stay: true })
                    break;
                case "frop":
                    // Frop
                    notie.alert({ type: 'info', text: data["text"], stay: true })
                    frop(data["cards"]);
                    break;
                case "turn":
                    // Turn
                    notie.alert({ type: 'info', text: data["text"], stay: true })
                    turn(data["card"]);
                    break;
                case "river":
                    // River
                    notie.alert({ type: 'info', text: data["text"], stay: true })
                    river(data["card"]);
                    break;
                case "result":
                    // Result
                    notie.alert({ type: 'info', text: data["text"], stay: true })
                    // Pop Up
                    Swal.fire({
                        icon: "info",
                        title: "Info",
                        text: "Show Down",
                        confirmButtonText: '<a href=' + data["URL"] + ' style="text-decoration:none; color:white; font-size:large; font-weight: bold;">Show Down</a>'
                    });
                    break;
                case "reset":
                    // Reset
                    console.log(jqXHRWBD)
                    stopAjax(jqXHRWBD, jqXHRWT)
                    document.location.href = data["redirect"];
                    break;
                case "popup":
                    // Pop Up
                    Swal.fire({
                        icon: "error",
                        title: "Something went wrong!",
                        text: "Please click OK",
                        confirmButtonText: '<a href="/remotepoker" style="text-decoration:none; color:white; font-size:large; font-weight: bold;">OK</a>'
                    });
                    break;
            }
            remoteWaitingPhase(jqXHRWT, jqXHRWBD);
        },
    });
    return jqXHRWT
}

function mobileWaitingBetData(jqXHRWBD, player, nowPlayer) {
    console.log("mobileWaitingBetData")
    jqXHRWBD = $.ajax({
        url: '/ajax/waitingpokerdata',
        dataType: 'json',
        success: function (data) {
            console.log(data)
            nowPlayer = data["decisionMaker"]
            $('#decisionMaker').text(nowPlayer)
            $('#betSize').text(data["betSize"]);
            $('#originalRaiser').text(data["originalRaiser"]);
            if (player === nowPlayer) {
                bettingbuttonsEnable()
                text = data["originalRaiser"] + " bet " + data["betSize"] + " dollars"
                popup("info", "Your Turn", text)
            } else {
                text = nowPlayer + " is playing"
                temppopup("info", "Info", text, "center", 1)
            }
            mobileWaitingBetData(jqXHRWBD, player, nowPlayer)
        },
    });
    return jqXHRWBD;
}

function frop(cards) {
    $.each(cards, function (index, card) {
        //カードを裏の画像から表の画像に切り替える
        $("#CommunityCard" + String(index)).attr('src', "/static/images/cardsIMG/" + card.Num + "_of_" + card.Suit + ".png");
    })
};

function turn(card) {
    //カードを裏の画像から表の画像に切り替える
    $("#CommunityCard3").attr('src', "/static/images/cardsIMG/" + card.Num + "_of_" + card.Suit + ".png");
};

function river(card) {
    //カードを裏の画像から表の画像に切り替える
    $("#CommunityCard4").attr('src', "/static/images/cardsIMG/" + card.Num + "_of_" + card.Suit + ".png");
};

function navResult(url) {
    // Pop Up
    Swal.fire({
        icon: "info",
        title: "Result",
        text: "Show Down",
        confirmButtonText: '<a href=' + url + ' style="text-decoration:none; color:white; font-size:large; font-weight: bold;">Result page</a>'
    });
};

function waitingPhase(jqXHR) {
    console.log("mobileWaitingPhase")
    jqXHR = $.ajax({
        url: '/ajax/waitingphase',
        dataType: 'json',
        success: function (data) {
            console.log(data)
            switch (data["func"]) {
                case "frop":
                    popup("info", "Info", data["text"]);
                    frop(data["cards"]);
                    break;
                case "turn":
                    popup("info", "Info", data["text"]);
                    turn(data["card"]);
                    break;
                case "river":
                    popup("info", "Info", data["text"]);
                    river(data["card"]);
                    break
                case "result":
                    navResult(data["URL"]);
                    break;
            }
            waitingPhase(jqXHR)
        },
    });
    return jqXHR
};

function bettingbuttonsDisable() {
    $(".bettingbuttons button").each(function (i, v) {
        $(v).prop("disabled", true);
    });
}

function bettingbuttonsEnable() {
    $(".bettingbuttons button").each(function (i, v) {
        $(v).prop("disabled", false);
    });
}

function remoteWaitForRedirect() {
    $.ajax({
        url: '/ajax/mobilewaitingphase',
        data: { from: "remotepoker" },
        dataType: 'json',
        success: async function (data) {
            switch (data["func"]) {
                case "reset":
                    // Reset
                    document.location.href = data["redirect"];
                    break;
                case "popup":
                    // Pop Up
                    Swal.fire({
                        icon: "error",
                        title: "Something went wrong!",
                        text: "Please click OK",
                        confirmButtonText: '<a href="/remotepoker" style="text-decoration:none; color:white; font-size:large; font-weight: bold;">OK</a>'
                    });
                    break;
            }
        },
    });
}

function mobileWaitForRedirect() {
    $.ajax({
        url: '/ajax/mobilewaitingphase',
        dataType: 'json',
        success: async function (data) {
            switch (data["func"]) {
                case "reset":
                    // Reset
                    document.location.href = data["redirect"];
                    break;
                case "popup":
                    // Pop Up
                    Swal.fire({
                        icon: "error",
                        title: "Something went wrong!",
                        text: "Please click OK",
                        confirmButtonText: '<a href="/mobilepoker" style="text-decoration:none; color:white; font-size:large; font-weight: bold;">OK</a>'
                    });
                    break;
            }
        },
    });
}