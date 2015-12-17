$module('functioncall', function () {
  var $instance = this;
  $instance.DoSomething = function () {
    var $state = {
      current: 0,
      returnValue: null,
    };
    var $returnValue$1;
    $state.next = function ($callback) {
      try {
        while (true) {
          switch ($state.current) {
            case 0:
              AnotherFunction(2).then(function (returnValue) {
                $state.current = 1;
                $returnValue$1 = returnValue;
                $state.next($callback);
              }).catch(function (e) {
                $state.error = e;
                $state.current = -1;
                $callback($state);
              });
              return;

            case 1:
              $returnValue$1;
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
  $instance.AnotherFunction = function (someparam) {
    return $promise.empty();
  };
});