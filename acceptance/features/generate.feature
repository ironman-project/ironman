
Feature: Generate a template
  In order to generate based on a template
  As a developer
  I need to run the generate command 

  Scenario: Generate a template with correct ID
    Given A template to generate installed with URL "https://github.com/ironman-project/template-example.git"
    When Generate runs with ID "template-example" generator ID "app" and flags "foo=bar,bar=foo"
    Then The generate process state should be success 
    And The generate output should contain "Running template generator app" and "done"
    And A file "foobar.txt" under the generated path should contain "Foo is bar, Bar is foo"
   
  Scenario: Generate a template with non existing ID
    When Generate with non existing id runs with ID "template-example-dont-exists" generator ID "app"
    Then The generate with non existing ID process state should be failure
    And The generate with non existing ID output should cointain "Template is not installed"