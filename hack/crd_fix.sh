#!/bin/bash
#
# A script to fix the bug in the controller

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

sed -i.bak  "130s/type: object//" $SCRIPTDIR/../config/crds/ibmcloud_v1alpha1_function.yaml
sed -i.bak  "100s/type: object//" $SCRIPTDIR/../config/crds/ibmcloud_v1alpha1_invocation.yaml
sed -i.bak  "46s/type: object//" $SCRIPTDIR/../config/crds/ibmcloud_v1alpha1_package.yaml
sed -i.bak  "88s/type: object//" $SCRIPTDIR/../config/crds/ibmcloud_v1alpha1_trigger.yaml
sed -i.bak  "5s/manager-rolebinding/cloud-functions-operator/" $SCRIPTDIR/../config/rbac/rbac_role_binding.yaml
sed -i.bak  "12s/default/cloud-functions-operator/" $SCRIPTDIR/../config/rbac/rbac_role_binding.yaml
sed -i.bak  "13s/system/ibmcloud-operators/" $SCRIPTDIR/../config/rbac/rbac_role_binding.yaml
sed -i.bak  "9s/manager-role/cloud-functions-operator/" $SCRIPTDIR/../config/rbac/rbac_role_binding.yaml
sed -i.bak  "5s/manager-role/cloud-functions-operator/" $SCRIPTDIR/../config/rbac/rbac_role.yaml

rm -f $SCRIPTDIR/../config/crds/*.bak
rm -f $SCRIPTDIR/../config/rbac/*.bak
