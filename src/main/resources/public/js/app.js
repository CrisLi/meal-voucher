$(function() {

  var user = {};
  var voucher = {};
  var vouchers = [];

  axios.get('/auth')
    .then(res => {

      user = res.data;

      appendUserInfo();

      if (user.role === 'ADMIN') {
        return hanldeAdmin();
      }

      return handleMember();
    })
    .catch(e => {

      console.log(e);

      if (_.get(e, 'response.status') === 401) {
        window.location.href = './login.html';
      }

    });

  $('#logout-btn').click(() => {
    axios.delete('/auth')
      .then(res => {
        window.location.href = './login.html';
      })
      .catch(e => {
        console.log(e);
      });
  });

  $('#change-voucher-count-btn').click(() => {

    var quota = $('#quota-input').val();

    if (parseInt(quota) && quota > 0) {
      axios.post(`/vouchers/${voucher.id}/change_quota?quota=${quota}`)
        .then(res => {
          alert('Change Successfully!');
          window.location.href = './';
        })
        .catch(e => {
          alert('Change Voucher Quota Failed!');
          console.log(e);
        });
    } else {
      alert('Please input a valid quota number!');
    }

  });

  function handleMember() {

    showEditRow();

    var teamId = user.team.id;
    var now = new Date();
    var year = now.getFullYear();
    var month = now.getMonth() + 1;

    return axios.get(`/vouchers/team?teamId=${teamId}&year=${year}&month=${month}`)
      .then(res => {
        voucher = res.data;
        appendVoucherInfo();
      });
  }

  function hanldeAdmin() {

    var now = new Date();
    var year = now.getFullYear();
    var month = now.getMonth() + 1;

    return axios.get(`/vouchers?year=${year}&month=${month}`)
      .then(res => {
        vouchers = res.data;
        appendVoucherList();
      });
  }

  function appendUserInfo() {
    $('#user-info').html(`Hello, <strong>${user.displayName}</strong>`);
  }

  function appendVoucherInfo() {
    $('#voucher-info').html(`Your team has ${voucher.quota} vocher(s) for this month (${voucher.month}/${voucher.year}).`);
    $('#voucher-list-row').hide();
    $('#voucher-info-row').show();
  }

  function appendVoucherList() {
    $('#voucher-list').html('');
    var list = $('<ul class="list-group" />');
    vouchers.forEach(v => {
      var item = $('<li class="list-group-item list-group-item-secondary" />')
        .html(`${v.team.name} - <strong>${v.quota}</strong> voucher(s) - ${v.month}/${v.year}`)
        .appendTo(list);
    });
    $('#voucher-list').append(list);
    $('#voucher-info-row').hide();
    $('#voucher-list-row').show();
  }

  function showEditRow() {
    if (user.role !== 'LEADER') {
      return;
    }
    var now = new Date();
    if (now.getDate() < 5) {
      $('#edit-row').show();
    }
  }

});
