# Roller

Roller is a small command line tool written in go to help working with a large number of AWS switch roles.

> Unless you work with a lot of AWS switch roles that change/extend often and are accessible in some form from a CMDB or the like, this is not much use to you.

Features:

* Easily add switch roles to the aws cli configuration.
* Clean up old switch roles which were added by Roller.
* Plugin based system to add loaders which can provide the switch roles.
* Autocompletion in zsh

It is aimed at users who (collaborately) work with a large number of AWS accounts and authenticate to those using switch roles.
The idea being that the roles are distributed to people in some form, but they would have to take care of setting them up in their CLI.
When it comes to 20+ roles, this may take some time when a new member joins the team. It can also get hard to keep up with new roles being added
or replaced with a differently named role as everyone would have to update their profiles individually.



## Setup

- Download the [binary](https://github.com/mitom/roller/releases) for your OS and put it on your `PATH` (make sure to rename it to `roller`!)
- Depending on your shell set up your rc file:
  - zsh:
    ```
    # ~/.zshrc
    roller completion zsh > "${fpath[1]}/_roller"
    eval $(roller init)
    ```
  - bash:
    ```
    # ~/.bashrc
    source <(roller completion bash)
    eval $(roller init)
    ```
- Configure a loader. Only a CSV loader is included, however loaders can be paths to Go plugins.
   ```
   # ~/.roller/config.yaml
   loader:
     my-loader:
       loader: csv
       ttl: 0
       options:
       path: ~/my-accounts.csv
   ```
- Create the source for the csv loader:
  ```
  <accountname>,<accountid>,<rolename>,<sessionduration>
  <accountname2>,<accountid2>,<rolename>,<sessionduration>
  <accountname>,<accountid>,<rolename2>,<sessionduration>
  ...
  ```
- launch a new shell and try to assume a role with `roller sw <tab><tab>` to see all the loaded accounts autocompleted.


## Example use

For the sake of these let's assume there is a role named `acc/role` loaded

- Assume a role: `roller sw acc/role`
- Open the switch role page in your browser `roller sw -w acc/role`
- Open the switch role page in your browser for the current role: `roller sw -w`
- Assume a role and give it an alias: `roller sw -n foo acc/role`
- Refresh your role _if needed_: `roller sw`
- List all roles loaded: `roller cache`
- Remove all expired sessions from the aws credentials file: `roller cleanup`


Roller has 2 ways of functioning:

1. The first one is essentially just providing all the information needed to switch to a role (account id and role name). This is not much benefit by itself.
2. The second one is using a cache loader. A cache loader is a plugin written in go that implements a simple interface. The purpose of it is to 
take an input (preferably a centralised place) where and parse out the information and return it. Roller than caches that information and provides autocompletion on it.
In the end what can be achieved is something along the lines of providing roller with a csv sheet like: `some random thing, aws account id, other random thing, RoleName`
Roller will parse it (with the right cache loader properly set up) and when the user goes to type in `roller switch [tab][tab]`, it will provide `aws-account-id/RoleName`
as an option to switch to. Selecting it will grab all the information from the cache needed to switch the role, and roller will generate temporary (1 hour) credentials via STS.

