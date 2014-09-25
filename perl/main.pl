#!/usr/bin/env perl
#
# Copyright 2014 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

use strict;
use warnings;

use Path::Class;
use Net::OAuth2::Client;

#
# This example gets the read bases for a sample at specific a position
#
my $dataset_id = 376902546192; # This is the 1000 Genomes dataset ID
my $sample = "NA12872";
my $reference_name = "22";
my $reference_position = 51003836;

my $token = get_access_token();

# 1. First find the readset ID for the sample
my $json = call_api($token, "POST",
    "readsets/search?fields=readsets(id)",
    ("datasetIds" => [$dataset_id], "name" => $sample));
my @readsets = @{$json->{readsets}};

scalar(@readsets) == 1 or die "Searching for " . $sample .
    " didn't return the right number of readsets";
my $readset_id = $readsets[0]->{id};


# 2. Once we have the readset ID,
# lookup the reads at the position we are interested in
$json = call_api($token, "POST",
    "reads/search?fields=reads(position,originalBases,cigar)",
    ("readsetIds" => [$readset_id],
     "sequenceName" => $reference_name,
     "sequenceStart" => $reference_position,
     "sequenceEnd" => $reference_position,
     "maxResults" => "1024"));

my %bases;
foreach (@{$json->{reads}}) {
  # Note: This is simplistic - the cigar should be considered for real code
  my $base = substr($_->{originalBases},
      $reference_position - $_->{position}, 1);
  $bases{$base}++;
}

print $sample, " bases on ", $reference_name, " at ",
    $reference_position, " are\n";
foreach my $base (keys %bases) {
  print "$base: $bases{$base}\n";
}


#
# This example gets the variants for a sample at a specific position
#

# 1. First find the call set ID for the sample
$json = call_api($token, "POST",
    "callsets/search?fields=callSets(id)",
    ("variantSetIds" => [$dataset_id], "name" => $sample));
my @call_sets = @{$json->{callSets}};

scalar(@call_sets) == 1 or die "Searching for " . $sample .
    " didn't return the right number of call sets";
my $call_set_id = $call_sets[0]->{id};


# 2. Once we have the call set ID,
# lookup the variants that overlap the position we are interested in
$json = call_api($token, "POST",
    "variants/search?fields=variants(names,referenceBases,alternateBases," .
    "calls(genotype))",
    ("callSetIds" => [$call_set_id],
     "referenceName" => $reference_name,
     # Note: currently, variants are 0-based and reads are 1-based,
     # reads will move to 0-based coordinates in the next version of the API
     "start" => $reference_position - 1,
     "end" => $reference_position));

my $variant = @{$json->{variants}}[0];
my $variant_name = $variant->{names}[0];

my @genotype;
foreach (@{@{$variant->{calls}}[0]->{genotype}}) {
  if ($_ == 0) {
    push(@genotype, $variant->{referenceBases});
  } else {
    push(@genotype, @{$variant->{alternateBases}}[$_ - 1]);
  }
}

print "the called genotype is ", @genotype, " for ", $variant_name, "\n";


# Authorization code
sub get_access_token {
  # Create a Client object from the client_secrets.json file in this directory
  my $client_secrets_data = file("client_secrets.json")->slurp();
  my $client_secrets_json = JSON->new->decode($client_secrets_data);

  my ($client_type) = keys(%$client_secrets_json);
  $client_secrets_json = $client_secrets_json->{$client_type};

  my $client = Net::OAuth2::Client->new(
    $client_secrets_json->{client_id},
    $client_secrets_json->{client_secret},
    authorize_path => $client_secrets_json->{auth_uri},
    access_token_path => $client_secrets_json->{token_uri},
    scope => "https://www.googleapis.com/auth/genomics"
  )->web_server(
    redirect_uri => @{$client_secrets_json->{redirect_uris}}[0]
  );

  # Load a previously saved access token, or start an OAuth authorization flow
  my $access_token_filename = "credentials.dat";

  if (-f $access_token_filename) {
    my $json = JSON->new->decode(file($access_token_filename)->slurp());
    $json->{client} = $client;
    return Net::OAuth2::AccessToken->new(%$json);

  } else {
    print "Go to the following link in your browser:\n\n",
        $client->authorize(), "\n\n", "Enter verification code: ";
    my $code = <STDIN>;
    my $access_token = $client->get_access_token($code);

    file($access_token_filename)->spew($access_token->to_string);
    return $access_token;
  }
}

# API helper method
sub call_api {
  my($token, $method, $path, %body) = @_;

  my $base_url = "https://www.googleapis.com/genomics/v1beta/";
  my $json_header = HTTP::Headers->new(Content_Type => "application/json");

  my $response = $token->request($method, $base_url . $path,
      $json_header, JSON->new->utf8->encode(\%body));

  return JSON->new->decode($response->decoded_content);
}