$module('compare', function () {
  var $instance = this;
  this.cls('SomeClass', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.$new = function () {
      var instance = new $static();
      function () {
      }.call(instance);
      return instance;
    };
  });
  $instance.DoSomething = function (first, second) {
    var $this = this;
    var $state = {
      current: 0,
      returnValue: null,
    };
    var $returnValue$1;
    var $returnValue$2;
    var $returnValue$3;
    var $returnValue$4;
    var $returnValue$5;
    var $returnValue$6;
    $state.next = function ($callback) {
      try {
        while (true) {
          switch ($state.current) {
            case 0:
              $g.compare.SomeClass.$equals(first, second).then(function (returnValue) {
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
              $g.compare.SomeClass.$equals(first, second).then(function (returnValue) {
                $state.current = 2;
                $returnValue$2 = returnValue;
                $state.next($callback);
              }).catch(function (e) {
                $state.error = e;
                $state.current = -1;
                $callback($state);
              });
              return;

            case 2:
              !$returnValue$2;
              $g.compare.SomeClass.$compare(first, second).then(function (returnValue) {
                $state.current = 3;
                $returnValue$3 = returnValue;
                $state.next($callback);
              }).catch(function (e) {
                $state.error = e;
                $state.current = -1;
                $callback($state);
              });
              return;

            case 3:
              $returnValue$3 < 0;
              $g.compare.SomeClass.$compare(first, second).then(function (returnValue) {
                $state.current = 4;
                $returnValue$4 = returnValue;
                $state.next($callback);
              }).catch(function (e) {
                $state.error = e;
                $state.current = -1;
                $callback($state);
              });
              return;

            case 4:
              $returnValue$4 > 0;
              $g.compare.SomeClass.$compare(first, second).then(function (returnValue) {
                $state.current = 5;
                $returnValue$5 = returnValue;
                $state.next($callback);
              }).catch(function (e) {
                $state.error = e;
                $state.current = -1;
                $callback($state);
              });
              return;

            case 5:
              $returnValue$5 <= 0;
              $g.compare.SomeClass.$compare(first, second).then(function (returnValue) {
                $state.current = 6;
                $returnValue$6 = returnValue;
                $state.next($callback);
              }).catch(function (e) {
                $state.error = e;
                $state.current = -1;
                $callback($state);
              });
              return;

            case 6:
              $returnValue$6 >= 0;
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