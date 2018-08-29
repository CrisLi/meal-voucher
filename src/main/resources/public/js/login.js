$(function() {
  $('#login-form').submit(function(e) {
    e.preventDefault();

    $('#login-btn').attr('disabled', 'disabled');

    var values = $(this).serializeArray();
    var json = {};

    values.forEach((e) => {
      json[e.name] = e.value;
    })

    if (json.login && json.password) {
      axios.post('/auth', json)
      .then(data => {
        window.location.href = "./index.html";
      })
      .catch(err => {
        console.error(err);
        alert('Login or password is wrong!');
        $('#login-btn').removeAttr('disabled');
      });
    } else {
      alert('Please input username and password');
      $('#login-btn').attr('disabled', 'disabled');
    }

  });
});
