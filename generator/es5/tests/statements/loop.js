
$module('loop', function() {
  var $instance = this;

  
  
  
  	
$instance.DoSomething = 

		function() {
			
			
			
				
	var $state = {
		current: 0,
		returnValue: null
	};

	

	$state.next = function($callback) {
		try {
			while (true) {
				switch ($state.current) {
					
					case 0:
						1234;

		$state.current = -1;
		return;
	

						break;
					
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
