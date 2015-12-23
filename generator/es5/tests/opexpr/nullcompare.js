$module('nullcompare', function () {
  var $static = this;
  $static.DoSomething = function (someParam) {
    var $state = {
      current: 0,
      returnValue: null,
    };
    $state.next = function ($callback) {
      try {
        while (true) {
          switch ($state.current) {
            case 0:
              $state.returnValue = $op.nullcompare(someParam, 2);
              $state.current = -1;
              $callback($state);
              return;

            default:
              $state.current = -1;
              return;
          }
        }
      } catch (e) {
        $state.error = e;
        $state.current = -1;
        $callback($state);
      }
    };
    return $promise.build($state);
  };
});
