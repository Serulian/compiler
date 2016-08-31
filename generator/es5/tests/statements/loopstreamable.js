$module('loopstreamable', function () {
  var $static = this;
  this.$class('SomeStreamable', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      var init = [];
      return $promise.all(init).then(function () {
        return instance;
      });
    };
    $instance.Stream = function () {
      var $this = this;
      var $result;
      var $current = 0;
      var $continue = function ($resolve, $reject) {
        while (true) {
          switch ($current) {
            case 0:
              $g.loopstreamable.SomeStream.new().then(function ($result0) {
                $result = $result0;
                $current = 1;
                $continue($resolve, $reject);
                return;
              }).catch(function (err) {
                $reject(err);
                return;
              });
              return;

            case 1:
              $resolve($result);
              return;

            default:
              $resolve();
              return;
          }
        }
      };
      return $promise.new($continue);
    };
    this.$typesig = function () {
      return $t.createtypesig(['Stream', 2, $g.____testlib.basictypes.Function($g.____testlib.basictypes.Stream($g.____testlib.basictypes.Boolean)).$typeref()], ['new', 1, $g.____testlib.basictypes.Function($g.loopstreamable.SomeStreamable).$typeref()]);
    };
  });

  this.$class('SomeStream', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      var init = [];
      init.push($promise.resolve($t.box(false, $g.____testlib.basictypes.Boolean)).then(function (result) {
        instance.wasChecked = result;
      }));
      return $promise.all(init).then(function () {
        return instance;
      });
    };
    $instance.Next = function () {
      var $this = this;
      var $result;
      var r;
      var $current = 0;
      var $continue = function ($resolve, $reject) {
        while (true) {
          switch ($current) {
            case 0:
              r = $this.wasChecked;
              $this.wasChecked = $t.box(true, $g.____testlib.basictypes.Boolean);
              $g.____testlib.basictypes.Tuple($g.____testlib.basictypes.Boolean, $g.____testlib.basictypes.Boolean).Build($t.box(true, $g.____testlib.basictypes.Boolean), $t.box(!$t.unbox(r), $g.____testlib.basictypes.Boolean)).then(function ($result0) {
                $result = $result0;
                $current = 1;
                $continue($resolve, $reject);
                return;
              }).catch(function (err) {
                $reject(err);
                return;
              });
              return;

            case 1:
              $resolve($result);
              return;

            default:
              $resolve();
              return;
          }
        }
      };
      return $promise.new($continue);
    };
    this.$typesig = function () {
      return $t.createtypesig(['Next', 2, $g.____testlib.basictypes.Function($g.____testlib.basictypes.Tuple($g.____testlib.basictypes.Boolean, $g.____testlib.basictypes.Boolean)).$typeref()], ['new', 1, $g.____testlib.basictypes.Function($g.loopstreamable.SomeStream).$typeref()]);
    };
  });

  $static.DoSomething = function (somethingElse) {
    var $result;
    var $temp0;
    var $temp1;
    var something;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      while (true) {
        switch ($current) {
          case 0:
            $t.box(1234, $g.____testlib.basictypes.Integer);
            $current = 1;
            continue;

          case 1:
            somethingElse.Stream().then(function ($result0) {
              $result = $result0;
              $current = 2;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 2:
            $temp1 = $result;
            $current = 3;
            continue;

          case 3:
            $temp1.Next().then(function ($result0) {
              $temp0 = $result0;
              $result = $temp0;
              $current = 4;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 4:
            something = $temp0.First;
            if ($t.unbox($temp0.Second)) {
              $current = 5;
              continue;
            } else {
              $current = 6;
              continue;
            }
            break;

          case 5:
            $t.box(7654, $g.____testlib.basictypes.Integer);
            $current = 3;
            continue;

          case 6:
            $t.box(5678, $g.____testlib.basictypes.Integer);
            $resolve();
            return;

          default:
            $resolve();
            return;
        }
      }
    };
    return $promise.new($continue);
  };
  $static.TEST = function () {
    var $result;
    var $temp0;
    var $temp1;
    var i;
    var result;
    var s;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      while (true) {
        switch ($current) {
          case 0:
            result = $t.box('noloop', $g.____testlib.basictypes.String);
            $g.loopstreamable.SomeStreamable.new().then(function ($result0) {
              $result = $result0;
              $current = 1;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 1:
            s = $result;
            $current = 2;
            continue;

          case 2:
            s.Stream().then(function ($result0) {
              $result = $result0;
              $current = 3;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 3:
            $temp1 = $result;
            $current = 4;
            continue;

          case 4:
            $temp1.Next().then(function ($result0) {
              $temp0 = $result0;
              $result = $temp0;
              $current = 5;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 5:
            i = $temp0.First;
            if ($t.unbox($temp0.Second)) {
              $current = 6;
              continue;
            } else {
              $current = 7;
              continue;
            }
            break;

          case 6:
            result = i;
            $current = 4;
            continue;

          case 7:
            $resolve(result);
            return;

          default:
            $resolve();
            return;
        }
      }
    };
    return $promise.new($continue);
  };
});
