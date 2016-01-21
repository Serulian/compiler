$module('streammember', function () {
  var $static = this;
  this.$class('SomeClass', false, function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      var init = [];
      init.push($promise.resolve(2).then(function (result) {
        instance.SomeInt = result;
      }));
      return $promise.all(init).then(function () {
        return instance;
      });
    };
  });

  $static.AnotherThing = function (somestream) {
    var $state = $t.sm(function ($callback) {
      while (true) {
        switch ($state.current) {
          case 0:
            $t.streamaccess(somestream, 'SomeInt');
            $state.current = -1;
            return;

          default:
            $state.current = -1;
            return;
        }
      }
    });
    return $promise.build($state);
  };
});
