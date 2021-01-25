function field_focus(field, email)
  {
    if(field.value == email)
    {
      field.value = '';
    }
  }

  function field_blur(field, email)
  {
    if(field.value == '')
    {
      field.value = email;
    }
  }

//Fade in dashboard box
$(document).ready(function(){
    $('.box').hide().fadeIn(1000);
    });

//Stop click event
$('a').click(function(event){
    event.preventDefault(); 
  });
$('div.btn').click(function(){
  $('form').submit();
});
$(document).keyup(function(event){
  if (event.keyCode==13){
    $('form').submit();
  }
})